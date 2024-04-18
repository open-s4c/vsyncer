// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os/exec"

	"vsync/checker"
	"vsync/logger"
)

type errorType int

//go:generate go run golang.org/x/tools/cmd/stringer -type=errorType
const (
	checkFail     errorType = 2
	internalError errorType = 1
	compilerError errorType = 1
	checkerError  errorType = 1
	noError       errorType = 0
)

type vError struct {
	typ    errorType
	status checker.CheckStatus
	err    error
}

func vfail(s checker.CheckStatus, err error) *vError {
	return &vError{
		typ:    checkFail,
		status: s,
		err:    err,
	}
}

func (e *vError) Error() string {
	switch e.typ {
	case checkFail:
		logger.Debugf("%v: %v", e.typ, e.status)
		return ""
	default:
		return e.err.Error()
	}
}

func (e *vError) Code() int {
	return int(e.typ)
}

func verror(typ errorType, err error) *vError {
	return &vError{
		typ: typ,
		err: err,
	}
}

type reportType string

func getErrorType(err error) string {
	if err == nil {
		return "none"
	}
	switch e := err.(type) {
	case *vError:
		return fmt.Sprintf("%v", e.typ)
	default:
		return "internalError"
	}
}

func getErrorCode(err error) int {
	if err == nil {
		return 0
	}
	switch e := err.(type) {
	case *vError:
		return e.Code()
	case *exec.ExitError:
		return e.ExitCode()
	default:
		return -1
	}
}

func getErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
