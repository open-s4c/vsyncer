// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package checker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/sync/errgroup"

	"vsync/logger"
	"vsync/tools"
)

func init() {
	tools.RegEnv("GENMC_CMD", "genmc", "Path to genmc binary")
	tools.RegEnv("GENMC_INCLUDE_PATH", "",
		"Path to genmc headers, e.g., genmc.h")
}

// GenMC is a wraps the GenMC model checker by Kokologiannakis et al.
type GenMC struct {
	threads uint
	mm      MemoryModel
	results []CheckResult
	version Version
}

var reExitStatus = regexp.MustCompile("^exit status [0-9]+$")

const genmcErrorCode = 42

func init() {
	tools.RegEnv("GENMC_OPTIONS", "",
		"Options passed to GenMC in additon to the default options")
	tools.RegEnv("GENMC_SET_OPTIONS", "",
		"Options passed to GenMC, replacing the default options")
}

// NewGenMC creates a new GenMC object
func NewGenMC(mm MemoryModel, threads uint) *GenMC {
	genmc := &GenMC{
		threads: threads,
		mm:      mm,
	}
	genmcCmd, err := tools.FindCmd("GENMC_CMD")
	if err != nil {
		logger.Fatalf("could not find genmc: %v", err)
	}
	genmc.setVersion(genmcCmd)
	return genmc
}

func (c *GenMC) setVersion(genmcCmd []string) {
	args := append(genmcCmd, "--version")
	my_ctx := context.Background()
	ostr, err := tools.RunCmdContext(my_ctx, args[0], args[1:], nil)
	if err != nil {
		logger.Fatalf("could not run genmc: %v", err)
	}
	r, err := regexp.Compile("v(\\d+)\\.(\\d+)(\\.(\\d+))?")
	if err != nil {
		logger.Fatalf("could not parse genmc version: %v", err)
	}
	grps := r.FindStringSubmatch(ostr)
	if len(grps) != 5 {
		logger.Fatalf("unexpected genmc version format: %v", grps)
	}
	c.version.major, _ = strconv.Atoi(grps[1])
	c.version.minor, _ = strconv.Atoi(grps[2])
	// group 3 is the optional dot so we skip it
	c.version.patch, _ = strconv.Atoi(grps[4])
	logger.Debugf("Detected GenMC version %d.%d.%d\n", c.version.major, c.version.minor, c.version.patch)
}

func (c *GenMC) GetVersion() string {
	return fmt.Sprintf("v%d.%d.%d", c.version.major, c.version.minor, c.version.patch)
}

func (c *GenMC) checkOne(ctx context.Context, genmcCmd []string, opts []string, i int) error {
	if len(c.results) <= i {
		return fmt.Errorf("unexpected index: %d", i)
	}

	cmd := genmcCmd[0]
	cmdArgs := append(genmcCmd[1:], opts...)
	logger.Debug(append([]string{cmd}, cmdArgs...))

	out, err := tools.RunCmdContext(ctx, cmd, cmdArgs, nil)
	if ctx.Err() == context.Canceled {
		return nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		// deadline reached, should be ok though
		c.results[i] = CheckResult{Status: CheckTimeout}
		return nil
	}
	fOutput := c.filterOutput(out)
	if err != nil {
		exiterr, ok := err.(*exec.ExitError)
		if !ok {
			return err
		}
		if exiterr.ExitCode() != genmcErrorCode {
			// in GenMC, all verification errors have exit code 42
			// internal errors have other exit codes
			logger.Debugf("Internal genmc error: %v", exiterr)

			msg := err.Error()
			match := reExitStatus.MatchString(msg)
			if !match && msg != "" {
				return fmt.Errorf("%s\n%v", out, exiterr)
			}
			return fmt.Errorf("%s", out)
		}
		if !c.doesTerminate(out) {
			c.results[i] = CheckResult{Status: CheckNotLive, Output: fOutput}
		} else {
			c.results[i] = CheckResult{Status: CheckNotSafe, Output: fOutput}
		}
		return nil
	}
	if !c.doesTerminate(out) {
		logger.Fatal("not live, but genmc gave no error status")
	}
	// extract the number of complete executions from the output
	r, regErr := regexp.Compile("Number of complete executions explored: (\\d+)\n")
	execNums := 0
	if regErr == nil {
		grps := r.FindStringSubmatch(fOutput)
		if len(grps) == 2 {
			execNums, regErr = strconv.Atoi(grps[1])
			logger.Debugf("Detected number of executions %d", execNums)
		}
	}
	// fail if there is no complete executions
	if execNums == 0 {
		// Problem with client code, zero executions explored
		text := `
Zero executions explored.
If your code uses __VERIFIER_assume(...), be sure you know what you are doing!`
		c.results[i] = CheckResult{Status: CheckRejected, Output: text, NumExecutions: 0}

	} else {
		c.results[i] = CheckResult{Status: CheckOK, Output: fOutput, NumExecutions: execNums}
	}
	logger.Infof("Genmc output: %s", fOutput)
	return nil
}

func (c *GenMC) getOpts() ([]string, error) {

	var extendedOpts []string

	// options for versions <= 0.9.X
	if c.version.major == 0 && c.version.minor < 8 {
		log.Fatal("VSyncer support GenMC v0.8.0 or higher")
	} else if c.version.major == 0 && c.version.minor >= 8 && c.version.minor < 10 {
		extendedOpts = []string{
			"-mo",
			"-check-liveness",
			"-disable-confirmation-annotation",
			"-disable-spin-assume",
			"-disable-load-annotation",
			"-disable-cast-elimination",
			"-disable-code-condenser",
		}
	} else {
		extendedOpts = []string{
			"-check-liveness",
			"-disable-estimation",
			"-disable-spin-assume",
		}
	}

	switch c.mm {
	case IMM:
		extendedOpts = append(extendedOpts, "-imm")

	case RC11:
		extendedOpts = append(extendedOpts, "-rc11")

	default:
		return nil, fmt.Errorf("genmc does not support '%v'", c.mm)
	}

	if env := tools.GetEnv("GENMC_OPTIONS"); env != "" {
		eopts := strings.Split(env, " ")
		extendedOpts = append(extendedOpts, eopts...)
	}

	if env := tools.GetEnv("GENMC_SET_OPTIONS"); env != "" {
		extendedOpts = strings.Split(env, " ")
	}
	return extendedOpts, nil
}

func (c *GenMC) checkResult(err error) (CheckResult, error) {
	if err != nil {
		logger.Debugf("===== genmc failed =====\n%v\n========================", err)
		return CheckResult{}, err
	}
	for _, r := range c.results {
		if r.Status == CheckNotLive || r.Status == CheckNotSafe || r.Status == CheckRejected {
			return r, nil
		}
	}
	for _, r := range c.results {
		if r.Status == CheckOK {
			return r, nil
		}
	}
	for _, r := range c.results {
		if r.Status == CheckTimeout {
			return r, nil
		}
	}
	logger.Fatal("should not get here")
	return CheckResult{}, nil
}

// Check runs GenMC on the module m
func (c *GenMC) Check(ctx context.Context, m DumpableModule) (cr CheckResult, err error) {
	fn, err := tools.Touch("input-*.ll")
	if err != nil {
		return cr, err
	}
	defer tools.Remove(fn)

	if err = tools.Dump(m, fn); err != nil {
		return cr, err
	}
	logger.Debug("checking", fn)

	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var genmcCmd []string
	// if the user did not specify a path use the environment var
	genmcCmd, err = tools.FindCmd("GENMC_CMD")
	if err != nil {
		return cr, err
	}

	extendedOpts, err := c.getOpts()
	if err != nil {
		return cr, err
	}

	optGroups := [][]string{append(extendedOpts, fn)}
	for i := uint(1); i < c.threads; i++ {
		opts := append(extendedOpts,
			fmt.Sprintf("-random-schedule-seed=%d", i),
			"-schedule-policy=random",
			fn)
		optGroups = append(optGroups, opts)
	}

	c.results = make([]CheckResult, len(optGroups))
	for i, opts := range optGroups {
		i, opts := i, opts
		g.Go(func() error {
			defer cancel()
			return c.checkOne(ctx, genmcCmd, opts, i)
		})
	}
	return c.checkResult(g.Wait())
}

func (c *GenMC) doesTerminate(str string) bool {
	return !strings.Contains(str, "Liveness violation!")
}

// filterOut filters the output of genmc to remove weird messages.
func (c *GenMC) filterOutput(out string) string {
	// remove anything before and including line with "Please submit"
	idx := strings.Index(out, "Please submit")
	if idx != -1 {
		lines := strings.Split(out[idx:], "\n")
		return strings.Join(lines[1:], "\n")
	}

	// if "Please submit" is not present, at least remove "^warning: .*$" lines
	var rlines []string
	for _, l := range strings.Split(out, "\n") {
		if !strings.Contains(l, "warning:") {
			rlines = append(rlines, l)
		}
	}
	return strings.Join(rlines, "\n")
}

func genMCIncludePaths() []string {
	// check if the user set the path for genmc includes, this is useful when
	//  --model-checker-path is used with checker
	if incPath := tools.GetEnv("GENMC_INCLUDE_PATH"); incPath != "" {
		logger.Debugf("GENMC_INCLUDE_PATH is set to=%s\n", incPath)
		if err := tools.FileExists(incPath); err != nil {
			logger.Fatalf("invalid genmc include path '%s': %v", incPath, err)
		}
		return []string{"-I", incPath}
	}

	// when GenMC runs with .ll file it prints the path the installed includes
	// so now we create an empty program, compile it to LLVM IR and call genmc
	const tinyProgram = `int main() { return 0; }`

	fn, err := tools.Touch("vsyncer-tiny-*.c")
	if err != nil {
		logger.Fatalf("could not create temporary file: %v", err)
	}
	defer os.Remove(fn)

	err = os.WriteFile(fn, []byte(tinyProgram), 0644)
	if err != nil {
		logger.Fatalf("could not write to temporary file: %v", err)
	}

	clangCmd, err := tools.FindCmd("CLANG_CMD")
	if err != nil {
		logger.Fatalf("could not find clang: %v", err)
	}

	var fnll = fn + ".ll"
	_, err = tools.RunCmd(clangCmd[0], append(clangCmd[1:],
		"-S", "-emit-llvm", "-o", fnll, fn), nil)
	if err != nil {
		logger.Fatalf("could not run clang: %v", err)
	}
	defer os.Remove(fnll)

	genmcCmd, err := tools.FindCmd("GENMC_CMD")
	if err != nil {
		logger.Fatalf("could not find genmc: %v", err)
	}

	output, err := tools.RunCmd(genmcCmd[0], append(genmcCmd[1:], "--", fnll), nil)
	if err != nil {
		logger.Fatalf("could not run genmc: %v", err)
	}

	paths := regexp.MustCompile(`'-I (.*)'`).FindAllStringSubmatch(output, -1)
	if len(paths) != 2 {
		logger.Fatalf("unexpected number of paths in genmc message: %v", paths)
	}

	// There must be 2 paths reported by GenMC:
	// - the path where GenMC was built
	// - the path where GenMC is supposed to be installed
	//
	// The result of FindAllStringSubmatch is a list of pairs:
	//   [ [complete-match, ()-group], ...]
	//
	// For example:
	//  p[0] = "'-I /some/path'"
	//  p[1] = "/some/path"
	//
	// We just want p[1].

	var incPaths []string
	for _, p := range paths {
		incPaths = append(incPaths, "-I", p[1])
	}
	if len(incPaths) == 0 {
		logger.Fatal("could not find any genmc include path")
	}
	return incPaths
}

func init() {
	compileOptions[GenmcID] =
		func() []string {
			return append(genMCIncludePaths(),
				"-D__CONFIG_GENMC_INODE_DATA_SIZE=64",
				"-DVSYNC_VERIFICATION_GENMC",
			)
		}
}
