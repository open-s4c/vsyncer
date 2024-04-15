// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package checker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"vsync/logger"
	"vsync/tools"
)

const fileMode = 0600

// DartagnanChecker wraps the Dartagnan model checker by Hernan Ponce de Leon et al.
type DartagnanChecker struct {
	mm MemoryModel
}

func init() {
	tools.RegEnv("JAVA_CMD", "java", "Path to java binary")

	tools.RegEnv("DARTAGNAN_HOME", "/usr/share/dat3m", "Path to DAT3M_HOME")
	tools.RegEnv("DARTAGNAN_OPTIONS", "",
		"Options passed to Dartagnan in additon to the default options")
	tools.RegEnv("DARTAGNAN_SET_OPTIONS", "",
		"Options passed to Dartagnan, replacing the default options")
}

// NewDartagnan creates a new checker using Dartagnan model checker.
func NewDartagnan(mm MemoryModel) *DartagnanChecker {
	return &DartagnanChecker{
		mm: mm,
	}
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
	GIMM:  {"genmc-imm.cat", "imm"},
	RC11:  {"rc11.cat", "c11"},
	VMM:   {"vmm.cat", "c11"},
}

func (c *DartagnanChecker) run(ctx context.Context, testFn string) (string, error) {
	dartagnanHome := tools.GetEnv("DARTAGNAN_HOME")

	opts := []string{
		"--property=program_spec,liveness",
		"--modeling.threadCreateAlwaysSucceeds=true",
		"--encoding.wmm.idl2sat=true",
		"--solver=yices2",
		fmt.Sprintf("--target=%s", models[c.mm].arch),
		fmt.Sprintf("%s/cat/%s", dartagnanHome, models[c.mm].cat),
	}

	if env := tools.GetEnv("DARTAGNAN_OPTIONS"); env != "" {
		opts = append(opts, strings.Split(env, " ")...)
	}

	if env := tools.GetEnv("DARTAGNAN_SET_OPTIONS"); env != "" {
		opts = strings.Split(env, " ")
	}

	args := append([]string{"-jar",
		dartagnanHome + "/dartagnan/target/dartagnan.jar",
		testFn,
	}, opts...)

	javaCmd, err := tools.FindCmd("JAVA_CMD")
	if err != nil {
		return "", err
	}
	logger.Debug(append(javaCmd, args...)) // just a message
	out, err := exec.CommandContext(ctx, javaCmd[0], append(javaCmd[1:], args...)...).CombinedOutput()
	return string(out), err
}

func (c *DartagnanChecker) GetVersion() string {
	// TODO: implement
	return "unknown"
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

	logger.Debug("Output:", sout)
	if err != nil {
		return cr, err
	}
	if strings.Contains(sout, "Program specification violation found") {
		return CheckResult{Status: CheckNotSafe, Output: sout}, nil
	} else if strings.Contains(sout, "Liveness violation found") {
		return CheckResult{Status: CheckNotLive, Output: sout}, nil
	} else if strings.Contains(sout, "CAT specification violation found") {
		return CheckResult{Status: CheckNotSafe, Output: sout}, nil
	} else if strings.Contains(sout, "Verification finished with result UNKNOWN\n") {
		text := `No violation found, but the program was not fully unrolled.
Try increasing the unrolling bound by adding "--bound=X" (where X is the bound) to DARTAGNAN_OPTIONS.`
		return CheckResult{Status: CheckRejected, Output: text}, nil
	} else if strings.Contains(sout, "Number of iterations: 1\n") {
		text := `Zero violating behaviors found.
Try increasing the unrolling bound by adding "--bound=X" (where X is the bound) to DARTAGNAN_OPTIONS.
If your code uses __VERIFIER_assume(...), be sure you know what you are doing!`
		return CheckResult{Status: CheckRejected, Output: text}, nil
	}
	return CheckResult{Status: CheckOK, Output: sout}, nil
}

func init() {
	compileOptions[DartagnanID] =
		func() []string {
			return []string{
				"-DVSYNC_VERIFICATION_DAT3M",
			}
		}
}
