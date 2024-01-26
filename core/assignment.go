// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

// Package core contains the most basic objects of vsyncer.
// These are assignment, bitsequences, selections, atomic operations and memory orderings.
package core

// Assignment is a pair of bitsequence and a selection
type Assignment struct {
	Bs  Bitseq
	Sel Selection
}
