// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"vsync/logger"
)

const fileMode = 0600

// Touch creates a new file temporary file with the given file pattern
func Touch(pattern string) (string, error) {
	tmp, err := ioutil.TempFile(".", pattern)
	if err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		logger.Warnf("error closing file: %v", err)
	}

	return ToSlash(tmp.Name()), nil
}

// RunCmd runs a command line with arguments and environment variable assignments
func RunCmd(cmdl string, args, env []string) (string, error) {
	return RunCmdContext(context.Background(), cmdl, args, env)
}

// RunCmdContext runs a command line with arguments and environment variable assignments and a context
func RunCmdContext(ctx context.Context, cmdl string, args, env []string) (string, error) {
	logger.Debug(append(append(env, cmdl), args...))
	cmd := exec.CommandContext(ctx, cmdl, args...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()

	sout := string(out)
	if err == nil {
		return sout, nil
	}
	if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
		return sout, nil
	}
	if err, ok := err.(*exec.Error); ok {
		return sout, err
	}
	if err, ok := err.(*exec.ExitError); ok {
		// remove newline of output
		if end := len(sout); end > 0 && sout[end-1] == '\n' {
			sout = sout[:len(sout)-1]
		}
		switch {
		case sout != "" && fmt.Sprintf("%v", err) != "exit status 1":
			return sout, err
		case sout != "":
			return sout, fmt.Errorf("%v", sout)
		default:
		}
		return sout, err
	}
	return sout, fmt.Errorf("unknown error: %v", err)
}

// MockFileExistsErr is a mock error returned by FileExists in tests
var MockFileExistsErr error

// FileExists returns nil if a file exists otherwise an error
func FileExists(fn string) error {
	if MockFileExistsErr != nil {
		return MockFileExistsErr
	}
	if _, err := os.Stat(fn); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", fn)
	}
	return nil
}

func filesExist(fns []string) error {
	for _, fn := range fns {
		if err := FileExists(fn); err != nil {
			return err
		}
	}
	return nil
}

// CopyFile copies src into dst. Returns nil upon no error
func CopyFile(src, dst string) error {
	logger.Infof("copying file: '%s' -> '%s'", src, dst)
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("could not read '%v': %v", src, err)
	}

	err = ioutil.WriteFile(dst, input, fileMode)
	if err != nil {
		return fmt.Errorf("could not create '%v': %v", dst, err)
	}
	return nil
}

const enableRemove = true

// Remove deletes as file. It can be disabled with the enableRemove flag in the source.
func Remove(fn string) error {
	logger.Debugf("Remove file '%s'", fn)
	if enableRemove {
		return os.Remove(fn)
	}
	return nil
}

// Dump writes the current state of the module to a file.
func Dump(m fmt.Stringer, fn string) error {
	logger.Debugf("Dump file '%s'", fn)
	out, err := os.OpenFile(fn,
		os.O_TRUNC|os.O_WRONLY|os.O_CREATE, fileMode)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			logger.Warnf("error closing file: %v", err)
		}
	}()
	_, err = fmt.Fprint(out, m)
	return err
}
