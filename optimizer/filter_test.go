// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package optimizer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"vsync/core"
)

func TestFilterSet(t *testing.T) {
	var (
		yes = core.MustFromString("0x0")
		no  = core.MustFromString("0x3")
		f   = make(filterSet)
	)
	f.Set(yes)
	assert.True(t, f[yes.ToBinString()])
	assert.False(t, f[no.ToBinString()])
}

func TestFilterDup(t *testing.T) {
	var (
		yes = core.MustFromString("0x1")
		no1 = core.MustFromString("0x0")
		no2 = core.MustFromString("0x3")
		f   = make(filterSet)
	)
	f.Set(yes)
	assert.True(t, f.Contains(yes, Dup))
	assert.False(t, f.Contains(no1, Dup))
	assert.False(t, f.Contains(no2, Dup))
}

func TestFilterRlx(t *testing.T) {
	var (
		yes  = core.MustFromString("0x1")
		also = core.MustFromString("0x0")
		no   = core.MustFromString("0x3")
		f    = make(filterSet)
	)
	f.Set(yes)
	assert.True(t, f.Contains(yes, Rlx))
	assert.True(t, f.Contains(also, Rlx))
	assert.False(t, f.Contains(no, Rlx))
}
