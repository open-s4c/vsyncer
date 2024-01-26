// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package checker

import "vsync/logger"

// MemoryModel is a set of constans identifying different memory models supported by some of the checkers.
//
//go:generate go run golang.org/x/tools/cmd/stringer -type=MemoryModel
type MemoryModel int

const (
	// InvalidMemoryModel represents any unknown memory model
	InvalidMemoryModel MemoryModel = iota
	// TSO memory model
	TSO
	// ARM8 memory model
	ARM8
	// Power memory model
	Power
	// RiscV memory model
	RiscV
	// IMM memory model
	IMM
	// GIMM memory model
	GIMM
	// RC11 memory model
	RC11
)

// ParseMemoryModel parses a string and returns an equivalent memory model identifier.
func ParseMemoryModel(mm string) MemoryModel {
	logger.Debugf("parsing memory model '%s'", mm)

	switch mm {
	case "tso":
		return TSO
	case "arm8":
		return ARM8
	case "power":
		return Power
	case "riscv":
		return RiscV
	case "imm":
		return IMM
	case "gimm":
		return GIMM
	case "rc11":
		return RC11
	default:
		return InvalidMemoryModel
	}
}
