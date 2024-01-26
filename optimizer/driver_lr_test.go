// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package optimizer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"vsync/checker"
	"vsync/core"
	"vsync/logger"
)

const u3 = 3

type mockChecker struct {
}

type mockModule struct {
	count  int
	oracle map[string]checker.CheckResult
}

func (m *mockModule) Bitseq(_ core.Selection) core.Bitseq { return core.Bitseq{} }
func (m *mockModule) Dump(_ string) error                 { return nil }
func (m *mockModule) Mutate(_ core.Assignment) error      { return nil }

func getClosure(m *mockModule, f filterSet) checkClosure {
	return func(ctx context.Context, bs core.Bitseq) (checker.CheckStatus, time.Duration) {
		m.count++
		s := m.oracle[bs.ToBinString()].Status
		if s == checker.CheckNotSafe || s == checker.CheckNotLive || s == checker.CheckInvalid {
			f.Set(bs)
		}
		return s, 0
	}
}

var (
	ctx = context.Background()
	d   = Driver{
		stats: NewStats(),
		cfg:   DriverConfig{Filter: Rlx},
	}
)

func TestDriverLrOK(t *testing.T) {
	d.filter = make(filterSet)
	m := &mockModule{
		oracle: map[string]checker.CheckResult{
			"0011": {Status: checker.CheckOK},
			"0000": {Status: checker.CheckNotSafe},
			"0010": {Status: checker.CheckNotLive},
			"0001": {Status: checker.CheckOK},
		},
	}

	sol := d.lr(ctx, core.MustFromString("0x3"), getClosure(m, d.filter))

	// there is one solution and that is 0001
	assert.True(t, len(sol) == 1)
	if len(sol) >= 1 {
		assert.Equal(t, sol[0].Bitseq().ToBinString(), "0001")
	}

	logger.Println(d.filter)

	// 0x3 was not checked
	assert.False(t, d.filter.Contains(core.MustFromString("0x3"), Rlx))

	// 0x0 was checked and failed
	assert.True(t, d.filter.Contains(core.MustFromString("0x0"), Rlx))

	// 0x1 was checked and succeeded
	assert.False(t, d.filter.Contains(core.MustFromString("0x1"), Rlx))

	// 0x2 was not checked
	assert.True(t, d.filter.Contains(core.MustFromString("0x2"), Rlx))

	// there were 3 checks
	assert.Equal(t, m.count, u3)
}

func TestDriverLrtimeout(t *testing.T) {
	d.filter = make(filterSet)
	m := &mockModule{
		oracle: map[string]checker.CheckResult{
			"0011": {Status: checker.CheckOK},
			"0000": {Status: checker.CheckNotSafe},
			"0010": {Status: checker.CheckNotLive},
			"0001": {Status: checker.CheckTimeout},
		},
	}

	sol := d.lr(ctx, core.MustFromString("0x3"), getClosure(m, d.filter))

	// there is one solution and that is 0001
	assert.True(t, len(sol) == 1)
	if len(sol) >= 1 {
		assert.Equal(t, sol[0].Bitseq().ToBinString(), "0001")
	}

	// 0x3 was not checked
	assert.False(t, d.filter.Contains(core.MustFromString("0x3"), Rlx))

	// 0x0 was checked and failed
	assert.True(t, d.filter.Contains(core.MustFromString("0x0"), Rlx))

	// 0x1 was checked and succeeded
	assert.False(t, d.filter.Contains(core.MustFromString("0x1"), Rlx))

	// 0x2 was checked and failed
	assert.True(t, d.filter.Contains(core.MustFromString("0x2"), Rlx))

	// there were 3 checks
	assert.Equal(t, u3, m.count)
}

func TestDriverLrnone(t *testing.T) {
	d.filter = make(filterSet)
	m := &mockModule{
		oracle: map[string]checker.CheckResult{
			"0011": {Status: checker.CheckOK},
			"0000": {Status: checker.CheckNotSafe},
			"0010": {Status: checker.CheckNotLive},
			"0001": {Status: checker.CheckNotSafe},
		},
	}

	sol := d.lr(ctx, core.MustFromString("0x3"), getClosure(m, d.filter))

	// there is no solution
	assert.True(t, len(sol) == 0)

	// 0x3 was not checked
	assert.False(t, d.filter.Contains(core.MustFromString("0x3"), Rlx))

	// 0x0, 0x1, and 0x2 were checked and failed
	assert.True(t, d.filter.Contains(core.MustFromString("0x0"), Rlx))
	assert.True(t, d.filter.Contains(core.MustFromString("0x1"), Rlx))
	assert.True(t, d.filter.Contains(core.MustFromString("0x2"), Rlx))

	// there were 3 checks
	assert.Equal(t, u3, m.count)
}
