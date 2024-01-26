// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"

	"vsync/core"
	"vsync/logger"
)

type wrapValues struct {
	ordering core.Ordering
	atomic   bool
}
type wrapInst struct {
	f      *ir.Func
	inst   ir.Instruction
	before wrapValues
	after  wrapValues
	stack  []meta
	id     int
}

func (w *wrapInst) wrapID() int { return w.id }

func newWrap(inst ir.Instruction, before wrapValues, f *ir.Func, s []meta, id int) wrapInst {
	if verboseVisitor {
		logger.Debugf("adding %d: %v", id, inst.LLString())
	}
	stack := make([]meta, len(s))
	copy(stack, s)
	return wrapInst{
		f:      f,
		inst:   inst,
		before: before,
		after:  before,
		stack:  stack,
		id:     id,
	}
}

func (w *wrapInst) setOrdering(o core.Ordering) {
	w.after.ordering = o
}

func (w *wrapInst) setAtomic(a bool) {
	w.after.atomic = a
}
func (w *wrapInst) isMutation() bool {
	return w.before != w.after
}
func (w *wrapInst) isAtomic(after bool) bool {
	if after {
		return w.after.atomic
	}
	return w.before.atomic
}
func (w *wrapInst) getOrdering(after bool) core.Ordering {
	if after {
		return w.after.ordering
	}
	return w.before.ordering
}
func (w *wrapInst) initialOrdering() core.Ordering {
	return w.before.ordering
}
func (w *wrapInst) wasAtomic() bool {
	return w.before.atomic
}
func (w *wrapInst) Mutate() (enum.AtomicOrdering, bool) {
	if !w.after.atomic {
		return toAtomicOrdering(core.Invalid), false
	}
	if w.after.ordering == core.Invalid {
		return toAtomicOrdering(core.SeqCst), true
	}
	return toAtomicOrdering(w.after.ordering), true
}
func (w *wrapInst) Unmutate() (enum.AtomicOrdering, bool) {
	return toAtomicOrdering(w.before.ordering), w.before.atomic
}

type wrapInstFence struct {
	*ir.InstFence
	wrapInst
}

func (w *wrapInstFence) LLString() string {
	if w.isMutation() {
		w.Ordering, _ = w.Mutate()
		defer func() {
			w.Ordering, _ = w.Unmutate()
		}()
	}
	if w.after.ordering == core.Relaxed {
		return ""
	}
	return w.InstFence.LLString()
}

type wrapInstLoad struct {
	*ir.InstLoad
	wrapInst
}

func (w *wrapInstLoad) LLString() string {
	if w.isMutation() {
		w.Ordering, w.Atomic = w.Mutate()
		defer func() {
			w.Ordering, w.Atomic = w.Unmutate()
		}()
	}
	return w.InstLoad.LLString()
}

type wrapInstStore struct {
	*ir.InstStore
	wrapInst
}

func (w *wrapInstStore) LLString() string {
	if w.isMutation() {
		w.Ordering, w.Atomic = w.Mutate()
		defer func() {
			w.Ordering, w.Atomic = w.Unmutate()
		}()
	}
	return w.InstStore.LLString()
}

type wrapInstAtomicRMW struct {
	*ir.InstAtomicRMW
	wrapInst
}

func (w *wrapInstAtomicRMW) LLString() string {
	if w.isMutation() {
		w.Ordering, _ = w.Mutate()
		defer func() {
			w.Ordering, _ = w.Unmutate()
		}()
	}
	return w.InstAtomicRMW.LLString()
}

type wrapInstCmpXchg struct {
	*ir.InstCmpXchg
	wrapInst
}

func (w *wrapInstCmpXchg) LLString() string {
	if w.isMutation() {
		w.SuccessOrdering, _ = w.Mutate()
		w.FailureOrdering = w.SuccessOrdering
		if w.FailureOrdering == enum.AtomicOrderingRelease {
			w.FailureOrdering = enum.AtomicOrderingMonotonic
		}
		defer func() {
			w.SuccessOrdering, _ = w.Unmutate()
			w.FailureOrdering = w.SuccessOrdering
			if w.FailureOrdering == enum.AtomicOrderingRelease {
				w.FailureOrdering = enum.AtomicOrderingMonotonic
			}
		}()
	}
	return w.InstCmpXchg.LLString()
}

type wrapInstruction interface {
	ir.Instruction
	isAtomic(after bool) bool
	setAtomic(a bool)
	getOrdering(after bool) core.Ordering
	setOrdering(o core.Ordering)
	diff() *diffEntry
}
