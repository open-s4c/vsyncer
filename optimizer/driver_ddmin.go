// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package optimizer

import (
	"context"

	"vsync/checker"
	"vsync/core"
)

func (d *Driver) ddmin2(ctx context.Context, bs core.Bitseq, check checkClosure, n int) []Solution {
	var bits = bs.Length()
	if bs.Ones() < n {
		return nil
	}

	var sd delta = bs.Indices()
	idxs := sd.Subslices(n)
	var deltas []core.Bitseq
	var nablas []core.Bitseq

	// check deltas
	for _, i := range idxs {
		delta := core.NewBitseq(bits).Set(i...)
		if !d.filter.Contains(delta, d.cfg.Filter) {
			deltas = append(deltas, delta)
		}
	}

	for _, sp := range deltas {
		status, _ := check(ctx, sp)
		if status == checker.CheckOK || status == checker.CheckTimeout {
			sol := d.ddmin2(ctx, sp, check, u2)
			return append(sol, Solution{bs: sp, status: status})
		}
		d.filter.Set(sp)
	}

	for _, i := range idxs {
		_ = i
		delta := core.NewBitseq(bits).Set(i...)
		nabla := bs.Xor(delta)
		if !d.filter.Contains(nabla, d.cfg.Filter) {
			nablas = append(nablas, nabla)
		}
	}

	for _, sp := range nablas {
		status, _ := check(ctx, sp)
		if status == checker.CheckOK || status == checker.CheckTimeout {
			sol := d.ddmin2(ctx, sp, check, max(n-1, u2))
			return append(sol, Solution{bs: sp, status: status})
		}
		d.filter.Set(sp)
	}
	if n < bs.Ones() {
		return d.ddmin2(ctx, bs, check, min(bs.Ones(), u2*n))
	}
	return nil
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
