// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

// Package checker contains a series of checkers that determine whether a given a module is correct, unsafe, or hangs.
package checker

import (
	"context"
)

// DumpableModule represents the interface of modules required by the checkers.
type DumpableModule interface {
	String() string
}

// Tool interface consists of one function to check the module and return a result.
type Tool interface {
	Check(ctx context.Context, m DumpableModule) (CheckResult, error)
	GetVersion() string
}

// CheckStatus represents the outcome of a check run
type CheckStatus int

//go:generate go run golang.org/x/tools/cmd/stringer -type=CheckStatus
const (
	// CheckUndefined represents a check with outcome Undefined
	CheckUndefined CheckStatus = iota
	// CheckOK represents a check with outcome OK
	CheckOK
	// CheckNotSafe represents a check with outcome NotSafe
	CheckNotSafe
	// CheckNotLive represents a check with outcome NotLive
	CheckNotLive
	// CheckInvalid represents a check with outcome Invalid
	CheckInvalid
	// CheckTimeout represents a check with outcome Timeout
	CheckTimeout
	// CheckRejected represents a check with outcome Rejected
	CheckRejected
)

// CheckResult is a pair of CheckStatus and output string
type CheckResult struct {
	Status        CheckStatus
	Output        string
	NumExecutions int
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=ID
type ID int

const (
	// Unknown checker
	UnknownID ID = iota
	// Dartagnan checker
	DartagnanID
	// GenMC checker
	GenmcID
	// Mock checker
	MockID
)

func ParseID(s string) ID {
	switch s {
	case "genmc":
		return GenmcID
	case "dartagnan":
		return DartagnanID
	case "mock":
		return MockID
	default:
		return UnknownID
	}
}

var compileOptions = map[ID][]string{}

func CompileOptions(id ID) []string {
	return compileOptions[id]

}
