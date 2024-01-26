// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package optimizer

import (
	"context"
	"time"

	"vsync/checker"
	"vsync/core"
)

const u2 = 2

func (d *Driver) lr(ctx context.Context, bs core.Bitseq, check checkClosure) []Solution {
	var sol []Solution
	for i := 0; i < bs.Length(); i += u2 {
		var seqs []core.Bitseq
		x := core.NewBitseq(bs.Length())

		if bs.Intersect(x.Set(i)) {
			seqs = append(seqs, bs.Unset(i))
		}
		if bs.Intersect(x.Set(i + 1)) {
			seqs = append(seqs, bs.Unset(i+1))
		}
		if len(seqs) == u2 {
			seqs = append([]core.Bitseq{bs.Unset(i, i+1)}, seqs...)
		}
		for _, s := range seqs {
			if d.filter.Contains(s, d.cfg.Filter) {
				continue
			}
			t := time.Now()
			status, _ := check(ctx, s)
			if status == checker.CheckOK || status == checker.CheckTimeout {
				bs = s
				sol = append(sol, Solution{bs: s, status: status, elapsed: time.Since(t)})
				break
			}

		}
	}
	reverseSolutions(sol)
	return sol
}

func reverseSolutions(s []Solution) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
