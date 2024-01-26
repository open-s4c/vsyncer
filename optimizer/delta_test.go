// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package optimizer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubsets(t *testing.T) {
	testCases := []struct {
		in  delta
		n   int
		out []delta
	}{
		{
			delta{0, 1, 2, 3, 4, 5, 6, 7}, 2,
			[]delta{{0, 1, 2, 3}, {4, 5, 6, 7}},
		}, {
			delta{0}, 2, nil,
		}, {
			delta{0, 1, 2, 3, 4, 5}, 4,
			[]delta{{0}, {1}, {2, 3}, {4, 5}},
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			o := tc.in.Subslices(tc.n)
			assert.Equal(t, tc.out, o)
		})
	}
}
