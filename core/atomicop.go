// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package core

//go:generate go run golang.org/x/tools/cmd/stringer -type=AtomicOp

// AtomicOp represents types of atomic operations
type AtomicOp int

const (
	// InvalidOp represents a InvalidOp operation
	InvalidOp AtomicOp = iota
	// Fence represents a Fence operation
	Fence
	// RMW represents a RMW operation
	RMW
	// Load represents a Load operation
	Load
	// Store represents a Store operation
	Store
	// Cmpxchg represents a Cmpxchg operation
	Cmpxchg
)

// GetOrdering returns the ordering of an atomic operation given a bit pair.
func (op AtomicOp) GetOrdering(val int) Ordering {
	return orderMap[op][val]
}
