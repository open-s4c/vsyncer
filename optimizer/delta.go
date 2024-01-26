// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package optimizer

import (
	"math/rand"
	"time"
)

type delta []int

func init() {
	rand.Seed(time.Now().UnixNano())
}
func (d delta) Subsets(n int) []delta {
	rand.Shuffle(n, func(i, j int) {
		if i < len(d) && j < len(d) {
			d[i], d[j] = d[j], d[i]
		}
	})

	return d.Subslices(n)
}

func (d delta) Subslices(n int) []delta {
	// via https://www.reddit.com/r/golang/comments/44cl7f/a_better_subslicing_algorithm/czp9r6j
	if n > len(d) {
		return nil
	}
	ret := make([]delta, 0, n)

	for ; n > 0; n-- {
		i := len(d) / n
		d, ret = d[i:], append(ret, d[:i])
	}

	return ret
}

func (d delta) ints() []int {
	var r []int
	for _, v := range d {
		r = append(r, v)
	}
	return r
}

func (d delta) remove(v int) delta {
	var ret delta
	for _, w := range d {
		if w != v {
			ret = append(ret, w)
		}
	}
	return ret
}

func complement(subsets []delta, n int) delta {
	var ret delta
	for i, d := range subsets {
		if i == n {
			continue
		}
		ret = append(ret, d...)
	}
	return ret
}
