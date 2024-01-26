// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package core

//go:generate go run golang.org/x/tools/cmd/stringer -type=Selection

// Selection represents a selection of operations of the target program
type Selection int

const (
	// SelectionInvalid does not select any operation
	SelectionInvalid Selection = iota
	// SelectionAtomic selects  rmws + fence + atomic loads + atomic stores
	SelectionAtomic
	// SelectionPlain selects plain loads + plain stores
	SelectionPlain
	// SelectionAtomicLoads selects AtomicLoads operations
	SelectionAtomicLoads
	// SelectionAtomicStores selects AtomicStores operations
	SelectionAtomicStores
	// SelectionRMWs selects RMWs operations
	SelectionRMWs
	// SelectionFences selects Fences operations
	SelectionFences
	// SelectionPlainLoads selects PlainLoads operations
	SelectionPlainLoads
	// SelectionPlainStores selects PlainStores operations
	SelectionPlainStores
	// SelectionLoads selects Loads operations
	SelectionLoads
	// SelectionStores selects Stores operations
	SelectionStores
)

// Group extracts sub selections of coarse selections
func (s Selection) Group() []Selection {
	var sel []Selection
	switch s {
	case SelectionAtomic:
		sel = append(sel, SelectionAtomicLoads, SelectionAtomicStores, SelectionFences, SelectionRMWs)
	case SelectionLoads:
		sel = append(sel, SelectionAtomicLoads, SelectionPlainLoads)
	case SelectionStores:
		sel = append(sel, SelectionAtomicStores, SelectionPlainStores)
	case SelectionPlain:
		sel = append(sel, SelectionPlainStores, SelectionPlainLoads)
	default:
		sel = append(sel, s)
	}
	return sel
}

// Binary returns whether the selection is about memory orderings (binary) or not
func (s Selection) Binary() bool {
	switch s {
	case SelectionLoads, SelectionStores, SelectionPlain:
		return false
	default:
		return true
	}
}
