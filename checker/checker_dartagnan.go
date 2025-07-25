// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package checker

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"path/filepath"
	"strings"

	"vsync/logger"
	"vsync/tools"
)

const fileMode = 0600

const BOUNDED_RESULT = 1
const PROGRAM_SPEC_VIOLATION = 10
const CAT_SPEC_VIOLATION = 11
const TERMINATION_VIOLATION = 12
const UNKNOWN_ERROR = 30

// DartagnanChecker wraps the Dartagnan model checker by Hernan Ponce de Leon et al.
type DartagnanChecker struct {
	mm MemoryModel
	version Version
}

func init() {
	tools.RegEnv("DARTAGNAN_JAVA_CMD", "java", "Path to java binary")

	tools.RegEnv("DARTAGNAN_HOME", "/usr/share/dat3m", "Path to DAT3M_HOME")
	tools.RegEnv("DARTAGNAN_OPTIONS", "",
		"Options passed to Dartagnan in additon to the default options")
	tools.RegEnv("DARTAGNAN_SET_OPTIONS", "",
		"Options passed to Dartagnan, replacing the default options")
	tools.RegEnv("DARTAGNAN_CAT_PATH", "", "Path to custom .cat files")
	tools.RegEnv("DARTAGNAN_SOLVER", "yices2", "Backend SMT solver (values: cvc4 | cvc5 | yices2 | z3)")
	tools.RegEnv("DARTAGNAN_BOUND", "", "Unroll bound integer (default unset)")
}

// NewDartagnan creates a new checker using Dartagnan model checker.
func NewDartagnan(mm MemoryModel) *DartagnanChecker {
	dartagnan := &DartagnanChecker{
		mm: mm,
	}
	dartagnan.setVersion()
	return dartagnan
}

func (c *DartagnanChecker) setVersion() {
	dartagnanHome := tools.GetEnv("DARTAGNAN_HOME")
	args := append([]string{"-jar",
		dartagnanHome + "/dartagnan/target/dartagnan.jar", "--version",
	})
	ctx := context.Background()
	javaCmd, err := tools.FindCmd("DARTAGNAN_JAVA_CMD")
	if err != nil {
		logger.Fatalf("could not run java: %v", err)
	}
	ostr, err := exec.CommandContext(ctx, javaCmd[0], append(javaCmd[1:], args...)...).CombinedOutput()
	if err != nil {
		logger.Fatalf("could not run dartagnan: %v", string(ostr))
	}
	r, err := regexp.Compile("(\\d+)\\.(\\d+)(\\.(\\d+))?")
	if err != nil {
		logger.Fatalf("could not parse dartagnan version: %v", err)
	}
	grps := r.FindStringSubmatch(string(ostr))
	if len(grps) != 5 {
		logger.Fatalf("unexpected dartagnan version format: %v", grps)
	}
	c.version.major, _ = strconv.Atoi(grps[1])
	c.version.minor, _ = strconv.Atoi(grps[2])
	// group 3 is the optional dot so we skip it
	c.version.patch, _ = strconv.Atoi(grps[4])
	logger.Debugf("Detected dartagnan version %d.%d.%d\n", c.version.major, c.version.minor, c.version.patch)
}

func (c *DartagnanChecker) GetVersion() string {
	return fmt.Sprintf("v%d.%d.%d", c.version.major, c.version.minor, c.version.patch)
}

var models = map[MemoryModel]struct {
	cat  string
	arch string
}{
	TSO:   {"tso.cat", "tso"},
	ARM8:  {"aarch64.cat", "arm8"},
	Power: {"power.cat", "power"},
	RiscV: {"riscv.cat", "riscv"},
	IMM:   {"imm.cat", "imm"},
	RC11:  {"rc11.cat", "c11"},
	VMM:   {"vmm.cat", "c11"},
}

func catFilePath(mm MemoryModel) string {
	dartagnanHome := tools.GetEnv("DARTAGNAN_HOME")

	modelInfo, has := models[mm]
	if !has {
		logger.Fatalf("could not find info for model '%v'", mm)
	}

	if b := tools.GetEnv("DARTAGNAN_CAT_PATH"); b != "" {
		if cpath := filepath.Join(b, modelInfo.cat); tools.FileExists(cpath) == nil {
			return tools.ToSlash(cpath)
		}
	}

	cpath := filepath.Join(dartagnanHome, "cat", modelInfo.cat)

	// we could be running dat3m in the container via "vsyncer docker". So, we should either:
	// A. return cpath even if it does not exist
	// B. check if we are running "vsyncer docker" and then check inside the container
	// For now, we go with option A.
	return tools.ToSlash(cpath)
}

func (c *DartagnanChecker) run(ctx context.Context, testFn string) (string, error) {

	opts := []string{
		"--encoding.wmm.idl2sat=true",
		"--bound.load=bound.csv",
		"--bound.save=bound.csv",
		fmt.Sprintf("--target=%s", models[c.mm].arch),
		catFilePath(c.mm),
	}

	if env := tools.GetEnv("DARTAGNAN_OPTIONS"); env != "" {
		opts = append(opts, strings.Split(env, " ")...)
	}

	if env := tools.GetEnv("DARTAGNAN_SET_OPTIONS"); env != "" {
		opts = strings.Split(env, " ")
	}

	if env := tools.GetEnv("DARTAGNAN_SOLVER"); env != "" {
		opts = append(opts, fmt.Sprintf("--solver=%s", env))
	}

	if env := tools.GetEnv("DARTAGNAN_BOUND"); env != "" {
		opts = append(opts, fmt.Sprintf("--bound=%s", env))
	}

	dartagnanHome := tools.GetEnv("DARTAGNAN_HOME")
	args := append([]string{"-jar",
		dartagnanHome + "/dartagnan/target/dartagnan.jar",
		testFn,
	}, opts...)

	javaCmd, err := tools.FindCmd("DARTAGNAN_JAVA_CMD")
	if err != nil {
		return "", err
	}
	logger.Debug(append(javaCmd, args...)) // just a message
	out, err := exec.CommandContext(ctx, javaCmd[0], append(javaCmd[1:], args...)...).CombinedOutput()
	return string(out), err
}

// Check performs a check run with Dartagnan
func (c *DartagnanChecker) Check(ctx context.Context, m DumpableModule) (cr CheckResult, err error) {
	testFn, err := tools.Touch("dartagnan-*.ll")
	if err != nil {
		return cr, err
	}
	defer tools.Remove(testFn)

	if err = tools.Dump(m, testFn); err != nil {
		return cr, err
	}
	sout, err := c.run(ctx, testFn)
	if ctx.Err() == context.Canceled {
		return cr, nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return CheckResult{Status: CheckTimeout}, nil
	}

	logger.Debug("Output:\n", sout)
	var result CheckResult
	if err != nil {
		exiterr := err.(*exec.ExitError)
		if exiterr.ExitCode() == BOUNDED_RESULT {
			logger.Debug("Increasing the unrolling bounds")
			result, _ = c.Check(ctx, m)
		}
		switch exiterr.ExitCode() {
			case PROGRAM_SPEC_VIOLATION, CAT_SPEC_VIOLATION:
				result = CheckResult{Status: CheckNotSafe, Output: sout}
			case TERMINATION_VIOLATION:
				result = CheckResult{Status: CheckNotLive, Output: sout}
			case UNKNOWN_ERROR:
				result = CheckResult{Status: CheckRejected, Output: sout}
		}
	} else {
		result = CheckResult{Status: CheckOK, Output: sout}
	}
	if strings.Contains(sout, "Number of iterations: 1\n") {
		text := `Zero violating behaviors found. If your code uses __VERIFIER_assume(...), be sure you know what you are doing!`
		result = CheckResult{Status: CheckRejected, Output: text}
	}
	tools.Remove("bound.csv")
	return result, nil
}

func init() {
	compileOptions[DartagnanID] =
		func() []string {
			return []string{
				"-DVSYNC_VERIFICATION_DAT3M",
				"-DVSYNC_DISABLE_SPIN_ANNOTATION",
			}
		}
}
