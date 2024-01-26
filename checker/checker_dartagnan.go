// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package checker

import (
	"context"
	"fmt"
	"os"
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
}

func (c *DartagnanChecker) prepare(m DumpableModule, testFn string) error {
	fn, err := tools.Touch("input-*.ll")
	if err != nil {
		return err
	}
	defer tools.Remove(fn)

	if err = tools.Dump(m, fn); err != nil {
		return err
	}

	obj, err := c.createVerifierMock()
	if err != nil {
		return err
	}
	defer tools.Remove(obj)

	llvmLink, err := tools.FindCmd("LLVM_LINK_CMD", "llvm-link")
	if err != nil {
		return fmt.Errorf("could not find llvm-link command: %v", err)
	}

	logger.Debug(append(llvmLink, fn, obj, "-S", "-o", testFn))
	out, err := exec.Command(llvmLink[0], append(llvmLink[1:], fn, obj, "-S", "-o", testFn)...).CombinedOutput()
	if err != nil {
		logger.Debug(string(out))
		return fmt.Errorf("could not link llvm files: %v", err)
	}
	return nil
}

func (c *DartagnanChecker) run(ctx context.Context, testFn string) (string, error) {
	dartagnanHome := "/dat3m"
	if env := os.Getenv("DARTAGNAN_HOME"); env != "" {
		dartagnanHome = env
	}

	opts := []string{
		"--property=program_spec,liveness",
		"--modeling.threadCreateAlwaysSucceeds=true",
		"--modeling.mixedType=true",
		"--modeling.precision=64",
		"--solver=yices2",
		fmt.Sprintf("--target=%s", models[c.mm].arch),
		fmt.Sprintf("%s/cat/%s", dartagnanHome, models[c.mm].cat),
	}

	if env := os.Getenv("DARTAGNAN_OPTIONS"); env != "" {
		opts = append(opts, strings.Split(env, " ")...)
	}

	if env := os.Getenv("DARTAGNAN_SET_OPTIONS"); env != "" {
		opts = strings.Split(env, " ")
	}

	args := append([]string{"-jar",
		dartagnanHome + "/dartagnan/target/dartagnan.jar",
		testFn,
	}, opts...)

	javaCmd, err := tools.FindCmd("DARTAGNAN_JAVA_CMD", "java")
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

	if err = c.prepare(m, testFn); err != nil {
		return cr, err
	}
	sout, err := c.run(ctx, testFn)
	if err != nil {
		return cr, err
	}

	logger.Debug(sout)
	if ctx.Err() == context.Canceled {
		return cr, nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return CheckResult{Status: CheckTimeout}, nil
	}
	if err != nil {
		return cr, err
	}
	if strings.Contains(sout, "Program specification violation found") {
		return CheckResult{Status: CheckNotSafe, Output: sout}, nil
	} else if strings.Contains(sout, "Liveness violation found") {
		return CheckResult{Status: CheckNotLive, Output: sout}, nil
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

func (c *DartagnanChecker) createVerifierMock() (string, error) {
	logger.Debug("compile verified mock")
	src, err := tools.Touch("verifier-mock-*.c")
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(src,
		os.O_TRUNC|os.O_WRONLY|os.O_CREATE, fileMode)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(f, dartagnanVerifierMock)
	if err := f.Close(); err != nil {
		return "", err
	}
	defer tools.Remove(f.Name())

	obj, err := tools.Touch("verifier-mock-*.ll")
	if err != nil {
		return "", err
	}

	clang, err := tools.FindCmd("CLANG_CMD", "clang")
	if err != nil {
		return "", fmt.Errorf("could not find clang commang: %v", err)
	}
	args := append(clang[1:], "-emit-llvm", "-S", "-o", obj, src)
	logger.Debug("calling", clang[0], args)
	out, err := exec.Command(clang[0], args...).CombinedOutput()
	if err != nil {
		logger.Debug(string(out))
		return "", fmt.Errorf("coult not execute clang command: %v", err)
	}
	return obj, nil
}

const dartagnanVerifierMock = `
#include <pthread.h>

int __VERIFIER_thread_create(pthread_t *thread, const pthread_attr_t *attr,
	void *(*start_routine) (void *), void *arg) {
	return pthread_create(thread, attr, start_routine, arg);
}
`
