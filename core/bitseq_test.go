// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"vsync/logger"
)

const u128 = 128

func FromUintAtLeast(v uint64) Bitseq {
	bs := FromUint(v)
	if bs.Length() < 64 {
		bs = bs.Fit(64)
	}
	return bs
}

func TestBitseqTrailing(t *testing.T) {
	testCases := []struct {
		v uint64
		c int
	}{
		{0, 64},
		{1, 0},
		{1 << 1, 1},
		{1 << 2, 2},
		{1 << 63, 63},
		{1<<63 + 1<<3, 3},
		{1<<63 + 1<<16 + 1<<3, 3},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			c := FromUintAtLeast(tc.v).Trailing()
			logger.Println(c)
			assert.Equal(t, tc.c, c)
		})
	}
}

func TestBitseqBinString(t *testing.T) {
	testCases := []struct {
		in  string
		out string
		err bool
	}{
		{in: "", err: true},
		{in: "00", out: "00"},
		{in: "01", out: "01"},
		{in: "1010101", out: "1010101"},
		{in: "10101010", out: "10101010"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			out, err := FromBinString(tc.in)
			if tc.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.out, out.ToBinString())
			}
		})
	}
}

func TestBitseqChild(t *testing.T) {
	testCases := []struct {
		v   uint64
		i   int
		c   uint64
		err bool
	}{
		{0x00, 0, 0x01, false},
		{0x00, 1, 0x02, false},
		{0x00, 63, 1 << 63, false},
		{0x00, 64, 1 << 63, true},
		{0x01, 0, 0, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			c, err := FromUintAtLeast(tc.v).Child(tc.i)
			if tc.err {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, FromUintAtLeast(tc.c), c)
			}
		})
	}
}

func TestBitseqAnd(t *testing.T) {
	testCases := []struct {
		v    uint64
		a, b int
		e    uint64
	}{
		{0xFF, 4, 8, 0xF0},
		{0xFF, 3, 8, 0xF8},
		{0xFFF, 3, 9, 0x1F8},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			e := FromUint(tc.v).Fit(u16).And(NewBitseq(u16).SetRange(tc.a, tc.b-1))
			assert.Equal(t, FromUint(tc.e).Fit(u16), e)
		})
	}
}

func TestBitseqSet(t *testing.T) {
	testCases := []struct {
		bits []int
		out  string
	}{
		{[]int{0, 1}, "11"},
		{[]int{2, 1}, "110"},
		{[]int{64}, "10000000000000000000000000000000000000000000000000000000000000000"},
		{[]int{65, 64, 1}, "110000000000000000000000000000000000000000000000000000000000000010"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			out := Bitseq{}.Set(tc.bits...)
			assert.Equal(t, tc.out, out.ToBinString())
		})
	}
}

func TestBitseqUnset(t *testing.T) {
	testCases := []struct {
		bits  []int
		bits2 []int
		out   string
	}{
		{[]int{0, 1}, []int{0, 1}, "00"},
		{[]int{2, 1}, []int{0, 1}, "100"},
		{[]int{64, 1}, []int{0, 1}, "10000000000000000000000000000000000000000000000000000000000000000"},
		{[]int{64, 1}, []int{0, 64}, "00000000000000000000000000000000000000000000000000000000000000010"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			out := Bitseq{}.Set(tc.bits...).Unset(tc.bits2...)
			assert.Equal(t, tc.out, out.ToBinString())
		})
	}
}

func TestBitseqSubsetOf(t *testing.T) {
	testCases := []struct {
		v1   []int
		v2   []int
		cond bool
	}{
		{[]int{0, 1}, []int{0, 1}, false},
		{[]int{2, 1}, []int{0, 1}, false},
		{[]int{2, 1}, []int{1}, true},
		{[]int{64, 1}, []int{0, 1}, false},
		{[]int{64, 1}, []int{1}, true},
		{[]int{124, 1}, []int{124}, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			cond := NewBitseq(u128).Set(tc.v2...).SubsetOf(
				NewBitseq(u128).Set(tc.v1...))
			assert.Equal(t, tc.cond, cond)
		})
	}
}

func TestBitseqRepeat(t *testing.T) {
	testCases := []struct {
		bits []int
		out  string
	}{
		{[]int{1}, "1100"},
		{[]int{0, 1}, "1111"},
		{[]int{2, 1}, "111100"},
		{[]int{3, 1}, "11001100"},
		{[]int{64, 1}, "110000000000000000000000000000000000000000000000" +
			"000000000000000000000000000000000000000000000000" +
			"0000000000000000000000000000001100"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			out := Bitseq{}.Set(tc.bits...).Repeat(u2)
			assert.Equal(t, tc.out, out.ToBinString())
		})
	}
}
