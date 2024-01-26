// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package optimizer

import (
	"vsync/core"
	"vsync/logger"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=filterStrategy

// filterStrategy determines how aggressive the filtering of filterSet can be applied.
type filterStrategy int

const (
	// None indicates no filtering
	None filterStrategy = iota
	// Dup  indicates a duplicate-based filtering
	Dup
	// Rlx indicates a relaxed filtering
	Rlx
)

type filterSet map[string]bool

func (fs filterSet) Set(bs core.Bitseq) {
	fs[bs.ToBinString()] = true
}

func (fs filterSet) Dup(bs core.Bitseq) bool {
	return fs[bs.ToBinString()]
}

func (fs filterSet) Rlx(bs core.Bitseq) bool {
	for so := range fs {
		so := core.MustFromBinString(so)
		if bs.SubsetOf(so) || bs.Equals(so) {
			return true
		}
	}
	return false
}

// Contains checks if the set contains the bitsequence using a filtering strategy.
func (fs filterSet) Contains(bs core.Bitseq, s filterStrategy) bool {
	switch s {
	case None:
		return false
	case Dup:
		return fs.Dup(bs)
	case Rlx:
		return fs.Rlx(bs)
	default:
		logger.Fatalf("unknown filter strategy %v", s)
	}
	return false
}
