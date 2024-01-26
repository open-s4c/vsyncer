// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"sort"

	"vsync/core"
	"vsync/logger"
)

type wrapInstSelection map[int]wrapInstruction

func (w wrapInstSelection) get(id int) wrapInstruction {
	return w[id]
}

func (w wrapInstSelection) sortedKeys() []int {
	var keys []int
	for k := range w {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (w wrapInstSelection) bitseqModes(after bool) core.Bitseq {
	var bs core.Bitseq
	bs = bs.Fit(len(w) * u2)
	for i, k := range w.sortedKeys() {
		var (
			in = w.get(k)
			i  = i * u2
			o  = in.getOrdering(after)
		)
		switch o {
		case core.Relaxed:
		case core.Release:
			bs = bs.Set(i)
		case core.Acquire:
			bs = bs.Set(i + 1)
		case core.SeqCst:
			bs = bs.Set(i, i+1)
		default:
			logger.Fatalf("mode not supported: %v", o)
		}
	}
	return bs
}

func (w wrapInstSelection) bitseqBinary(after bool) core.Bitseq {
	var bs core.Bitseq
	bs = bs.Fit(len(w))
	for i, k := range w.sortedKeys() {
		if w.get(k).isAtomic(after) {
			bs = bs.Set(i)
		}
	}
	return bs
}

func (w wrapInstSelection) Bitseq(modes bool, after bool) core.Bitseq {
	if modes {
		return w.bitseqModes(after)
	}
	return w.bitseqBinary(after)
}
