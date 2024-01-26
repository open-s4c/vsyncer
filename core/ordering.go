// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package core

//go:generate go run golang.org/x/tools/cmd/stringer -type=Ordering

// Ordering represents the memory ordering of an atomic operation
type Ordering int

const (
	// Invalid  memory ordering
	Invalid Ordering = iota
	// SeqCst memory ordering
	SeqCst
	// Acquire memory ordering
	Acquire
	// Release memory ordering
	Release
	// Relaxed memory ordering
	Relaxed
)

var (
	map4 = map[int]Ordering{
		0b00: Relaxed,
		0b01: Release,
		0b10: Acquire,
		0b11: SeqCst,
	}

	orderMap = map[AtomicOp]map[int]Ordering{
		Fence:   map4,
		RMW:     map4,
		Cmpxchg: map4,
		Load: {
			0b00: Relaxed,
			0b10: Acquire,
			0b11: SeqCst,
		},
		Store: {
			0b00: Relaxed,
			0b01: Release,
			0b11: SeqCst,
		},
	}
)
