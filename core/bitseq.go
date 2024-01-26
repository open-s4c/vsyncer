// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package core

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"vsync/logger"
)

// Bitseq represents a series of memory orderings or atomicity of operations as a sequence of bits
type Bitseq struct {
	data []uint64
	bits int
}

var (
	hexRe = regexp.MustCompile("^0x[0-9A-Fa-f]+$")
	binRe = regexp.MustCompile("^0b[01]+$")
)

const (
	u64 = 64
	u32 = 32
	u16 = 16
	u8  = 8
	u4  = 4
	u2  = 2
	u1  = 1
)

// ParseBitseq parses a bitsequence from a string with given bit length.
func ParseBitseq(str string, length int) (Bitseq, error) {
	switch str {
	case "0":
		return NewBitseq(length), nil
	case "-1":
		return NewBitseq(0).SetRange(0, length-1), nil
	default:
		s, err := FromString(str)
		if err != nil {
			return s, err
		}
		return s.Fit(length), nil
	}
}

// FromString parses a bit sequence from a string.
func FromString(str string) (Bitseq, error) {
	var s Bitseq
	switch {
	case hexRe.MatchString(str):
		return FromHexString(str[u2:])

	case binRe.MatchString(str):
		return FromBinString(str[u2:])

	default:
		return s, fmt.Errorf("cannot parse bitseq %v", str)
	}
}

// MustFromString parses a bit sequence from a string and abort the program on error.
func MustFromString(str string) Bitseq {
	s, err := FromString(str)
	if err != nil {
		logger.Fatal(err)
	}
	return s
}

// MustFromBinString parses a bit sequence from a string of '1's and '0's and abort the program on error.
func MustFromBinString(str string) Bitseq {
	s, err := FromBinString(str)
	if err != nil {
		logger.Fatal(err)
	}
	return s
}

// Length returns the number of bits of the bitsequence
func (s Bitseq) Length() int {
	return s.bits
}

// FromBinString parses a string assuming a binary format
func FromBinString(str string) (Bitseq, error) {
	if len(str) == 0 {
		return Bitseq{}, errors.New("cannot create empty bitseq")
	}
	var (
		s = Bitseq{
			data: make([]uint64, divceil(len(str), u64)),
			bits: len(str),
		}
		d = -1
	)
	for i := 0; i < len(str); i++ {
		if i%u64 == 0 {
			d++
		}
		if str[len(str)-i-1] == '1' {
			s.data[d] |= (1 << (i % u64))
		}
	}

	return s, nil
}

// FromHexString parses a string assuming hex format
func FromHexString(str string) (Bitseq, error) {
	if len(str) == 0 {
		return Bitseq{}, errors.New("cannot create empty bitseq")
	}
	var (
		length = divceil(len(str), u16) // 16 nibbles per uint64
		bs     Bitseq
	)
	for i := 0; i < length-1; i++ {
		l := u16
		i := len(str) - (i * u16) - u16
		d := str[i : i+l]
		u, err := strconv.ParseUint(d, u16, u64)
		if err != nil {
			return Bitseq{}, err
		}
		bs.data = append(bs.data, u)
	}
	l := len(str) - (length-1)*u16
	d := str[0:l]
	u, err := strconv.ParseUint(d, u16, u64)
	if err != nil {
		return Bitseq{}, err
	}
	bs.data = append(bs.data, u)
	bs.bits = len(str) * u4
	return bs, nil
}

func divceil(a, b int) int {
	return (a + b - 1) / b
}

// NewBitseq returns a new Bitseq object
func NewBitseq(bits int) Bitseq {
	sz := divceil(bits, u64)
	return Bitseq{
		bits: bits,
		data: make([]uint64, sz),
	}
}

// FromUint creates a bitsequence from an integer
func FromUint(v uint64) Bitseq {
	bstr := fmt.Sprintf("0b%b", v)
	return MustFromString(bstr)
}

// ToBinString returns the binary string representation
func (s Bitseq) ToBinString() string {
	var str string
	for i, parts := 0, divceil(s.bits, u64); i < parts && i < len(s.data); i++ {
		s := strconv.FormatUint(s.data[i], u2)
		for i < parts-1 && len(s) < u64 {
			s = "0" + s
		}
		str = s + str
	}
	for len(str) < s.bits {
		str = "0" + str
	}
	return str
}

// ToHexString returns the hex string representation
func (s Bitseq) ToHexString() string {
	var str string
	for i, parts := 0, divceil(s.bits, u64); i < parts && i < len(s.data); i++ {
		s := strconv.FormatUint(s.data[i], u16)
		for i < parts-1 && len(s) < u16 {
			s = "0" + s
		}
		str = s + str
	}
	for len(str) < divceil(s.bits, u4) {
		str = "0" + str
	}
	return fmt.Sprintf("%s", str)
}

func (s Bitseq) String() string {
	if s.Length() <= u32 {
		return fmt.Sprintf("0x%s (0b%s)", s.ToHexString(), s.ToBinString())
	}
	return fmt.Sprintf("0x%s", s.ToHexString())
}

// Equals returns true if s is equal to o.
func (s Bitseq) Equals(o Bitseq) bool {
	if s.bits != o.bits {
		return false
	}
	for i := range s.data {
		if i >= len(o.data) { // Make linter happy
			continue
		}
		if s.data[i] != o.data[i] {
			return false
		}
	}
	return true
}

// IsZero returns true if all bits are 0.
func (s Bitseq) IsZero() bool {
	for _, d := range s.data {
		if d != 0 {
			return false
		}
	}
	return true
}

// Fit extends or shrinks the number of bits of the Bitseq.
func (s Bitseq) Fit(n int) Bitseq {
	if n == 0 {
		s.bits = 0
		s.data = nil
	}
	if n < s.bits {
		s.bits = n
		s.data = s.data[0:divceil(s.bits, u64)]
		s.data[len(s.data)-1] &= (1 << (s.bits % u64)) - 1
	}
	if n > s.bits {
		s.bits = n
		c := divceil(s.bits, u64)
		for len(s.data) < c {
			s.data = append(s.data, 0)
		}
	}
	return s
}

// Set sets the i-th bit of the Bitseq extending it if necessary.
func (s Bitseq) Set(i ...int) Bitseq {
	s = s.Clone()
	for _, b := range i {
		if b >= s.bits {
			s = s.Fit(b + 1)
		}
		idx := b / u64
		if idx >= len(s.data) { // Make linter happy
			continue
		}
		s.data[idx] |= 1 << (b % u64)
	}
	return s
}

// SetRange sets a range [from; to] bits extending if necessary.
func (s Bitseq) SetRange(from, to int) Bitseq {
	s = s.Clone()
	for i := from; i <= to; i++ {
		s = s.Set(i)
	}
	return s
}

// Clone copies the bitsequence cloning the internal slice
func (s Bitseq) Clone() Bitseq {
	bs := NewBitseq(s.Length())
	copy(bs.data, s.data)
	return bs
}

// Unset unsets the i-th bit of the Bitseq extending it if necessary.
func (s Bitseq) Unset(i ...int) Bitseq {
	s = s.Clone()
	for _, b := range i {
		if b >= s.bits {
			s = s.Fit(b + 1)
		}
		idx := b / u64
		if idx >= len(s.data) { // Make linter happy
			continue
		}
		s.data[idx] &^= 1 << (b % u64)
	}
	return s
}

// SubsetOf tests whether s is a strict subset of o.
func (s Bitseq) SubsetOf(o Bitseq) bool {
	if len(s.data) != len(o.data) || s.bits != o.bits {
		return false
	}
	if s.Equals(o) {
		return false
	}
	for i := range s.data {
		if s.data[i]&o.data[i] != s.data[i] {
			return false
		}
	}
	return true
}

// Intersect returns true if s and o have 1 bits in common.
func (s Bitseq) Intersect(o Bitseq) bool {
	if len(s.data) != len(o.data) || s.bits != o.bits {
		return false
	}
	for i := range s.data {
		if s.data[i]&o.data[i] != 0 {
			return true
		}
	}
	return false
}

// Ones returns the number of one-bits in the Bitseq.
func (s Bitseq) Ones() int {
	var r int

	for _, d := range s.data {
		for ; d != 0; d >>= 1 {
			if d&1 == 1 {
				r++
			}
		}
	}
	return r
}

// Indices returns the indices of 1-bits.
func (s Bitseq) Indices() []int {
	var r []int
	for d, bs := range s.data {
		for i := 0; bs > 0; bs >>= 1 {
			if bs&1 == 1 {
				r = append(r, d*u64+i)
			}
			i++
		}
	}
	return r
}

// Repeat repeats the bits n times
func (s Bitseq) Repeat(n int) Bitseq {
	bs := NewBitseq(s.bits * n)
	for _, i := range s.Indices() {
		for j := i * n; j < (i+1)*n; j++ {
			bs = bs.Set(j)
		}
	}
	return bs
}

// And performs an And operation between two bitsequences and returns the result
func (s Bitseq) And(o Bitseq) Bitseq {
	if s.bits != o.bits {
		logger.Fatal("cannot AND different sized bitseqs")
	}
	var bs Bitseq
	for i := range s.data {
		if i >= len(o.data) { // Make linter happy
			continue
		}
		bs.data = append(bs.data, s.data[i]&o.data[i])
	}
	bs.bits = s.bits
	return bs
}

// Xor performs an Xor operation between two bitsequences and returns the result
func (s Bitseq) Xor(o Bitseq) Bitseq {
	if s.bits != o.bits {
		logger.Fatal("cannot XOR different sized bitseqs")
	}
	var bs Bitseq
	for i := range s.data {
		if i >= len(o.data) { // Make linter happy
			continue
		}
		bs.data = append(bs.data, s.data[i]^o.data[i])
	}
	bs.bits = s.bits
	return bs
}

// Or performs an Or operation between two bitsequences and returns the result
func (s Bitseq) Or(o Bitseq) Bitseq {
	if s.bits != o.bits {
		logger.Fatal("cannot OR different sized bitseqs")
	}
	var bs Bitseq
	for i := range s.data {
		if i >= len(o.data) { // Make linter happy
			continue
		}
		bs.data = append(bs.data, s.data[i]|o.data[i])
	}
	bs.bits = s.bits
	return bs
}

// Reverse reverses the order of the bits of a bitsequence maintaining its length
func (s Bitseq) Reverse() Bitseq {
	var ones []int
	for _, i := range s.Indices() {
		ones = append(ones, s.Length()-i-1)
	}
	return NewBitseq(s.Length()).Set(ones...)
}

// Trailing returns the number of trailing zeros in a Bitseq.
func (s Bitseq) Trailing() int {
	var count int

	if s.IsZero() {
		return s.bits
	}

	if s.data[0]&1 == 1 {
		return 0
	}
	for _, d := range s.data {
		for d != 0 && d&1 == 0 {
			count++
			d >>= 1
		}
		if d&1 != 0 {
			break
		}
	}

	return count
}

// Child returns the i-th child of s.
func (s Bitseq) Child(i int) (Bitseq, error) {
	count := s.Trailing()
	if i > count-1 {
		return Bitseq{}, fmt.Errorf("bitseq 0x%x has no %d-th child", s, i)
	}
	return s.Set(i), nil
}

// BitseqTranslate is a callback for the Translate function.
type BitseqTranslate func(k int, value int) error

// Translate can be used to convert a bitsequence to an arbitrary data oject using the translate closure.
func (s Bitseq) Translate(granularity int, translate BitseqTranslate) error {
	if granularity == 0 {
		return fmt.Errorf("cannot use empty mapping")
	}
	if s.bits%granularity != 0 {
		return fmt.Errorf("granularity (%v) is not multiple of bitseq length (%d)", granularity, s.bits)
	}
	if granularity > u2 {
		return fmt.Errorf("bitseq has not been tested with granularity %d", granularity)
	}

	var (
		sz    = u64 / granularity
		data  uint64
		mask  uint64
		items = s.bits / granularity
	)
	for i := 0; i < granularity; i++ {
		mask <<= u1
		mask |= u1
	}

	if items/sz > len(s.data) {
		return fmt.Errorf("unexpected number of items: %d", items)
	}

	for k := 0; k < items; k++ {
		if k%sz == 0 {
			data = s.data[k/sz]
		}
		val := int(data & mask)
		if err := translate(k, val); err != nil {
			return err
		}
		data >>= granularity
	}
	return nil
}
