// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/metadata"

	"vsync/core"
	"vsync/logger"
)

func (m *wrapModule) Assignment(sel core.Selection) core.Assignment {
	return core.Assignment{
		Bs:  m.bitseq(sel, true),
		Sel: sel,
	}
}

func getDbg(md meta) *metadata.Attachment {
	for _, m := range md.MDAttachments() {
		if m.Name == "dbg" {
			return m
		}
	}
	return nil
}

type analyzer struct {
	isDeclared map[interface{}]bool
	isAlloca   map[interface{}]bool
	isParam    map[interface{}]bool
	count      int
	mod        *wrapModule
}

func newAnalyzer(mod *wrapModule) *analyzer {
	return &analyzer{
		isDeclared: make(map[interface{}]bool),
		isAlloca:   make(map[interface{}]bool),
		isParam:    make(map[interface{}]bool),
		count:      0,
		mod:        mod,
	}
}
func (a *analyzer) instCall(inst *ir.InstCall) {
	// collect local variables that have a name
	if strings.Contains(inst.Callee.Ident(), "llvm.dbg.declare") ||
		strings.Contains(inst.Callee.Ident(), "llvm.dbg.addr") {
		op := *inst.Operands()[1]
		mv, ok := op.(*metadata.Value)
		if !ok {
			logger.Fatalf("could not case op %v", op)
		}
		a.isDeclared[mv.Value] = true
	}
}

func (a *analyzer) instLoad(i ir.Instruction, inst *ir.InstLoad, f *ir.Func, stack []meta) ir.Instruction {
	if getDbg(i.(meta)) == nil {
		return nil
	}
	// only loads that access local or global variables
	if src := inst.Src; true {
		if !a.isDeclared[src] && a.isAlloca[src] {
			return nil
		}
		if a.isParam[src] {
			return nil
		}
	}

	values := wrapValues{
		ordering: fromAtomicOrdering(inst.Ordering),
		atomic:   inst.Atomic,
	}

	in := &wrapInstLoad{InstLoad: inst, wrapInst: newWrap(inst, values, f, stack, a.count)}
	a.count++
	a.mod.addInst(a.count, in)
	return in

}

func (a *analyzer) instStore(i ir.Instruction, inst *ir.InstStore, f *ir.Func, stack []meta) ir.Instruction {
	// if a param is written to an alloca, then we can ignore that alloca
	// even if it is declared
	if getDbg(i.(meta)) == nil {
		if _, yes := inst.Src.(*ir.Param); yes {
			a.isParam[inst.Dst] = true
		}
		return nil
	}

	// only stores that access local or global variables
	if dst := inst.Dst; true {
		if !a.isDeclared[dst] && a.isAlloca[dst] {
			return nil
		}
		if a.isParam[dst] {
			return nil
		}
	}

	values := wrapValues{
		ordering: fromAtomicOrdering(inst.Ordering),
		atomic:   inst.Atomic,
	}

	in := &wrapInstStore{InstStore: inst, wrapInst: newWrap(inst, values, f, stack, a.count)}
	a.count++
	a.mod.addInst(a.count, in)
	return in
}

func (a *analyzer) instFence(inst *ir.InstFence, f *ir.Func, stack []meta) ir.Instruction {
	values := wrapValues{
		ordering: fromAtomicOrdering(inst.Ordering),
		atomic:   true,
	}
	in := &wrapInstFence{InstFence: inst, wrapInst: newWrap(inst, values, f, stack, a.count)}
	a.count++
	a.mod.addInst(a.count, in)
	return in
}

func (a *analyzer) instCmpXchg(inst *ir.InstCmpXchg, f *ir.Func, stack []meta) ir.Instruction {
	values := wrapValues{
		ordering: fromAtomicOrdering(inst.SuccessOrdering),
		atomic:   true,
	}
	in := &wrapInstCmpXchg{InstCmpXchg: inst, wrapInst: newWrap(inst, values, f, stack, a.count)}
	a.count++
	a.mod.addInst(a.count, in)
	return in
}

func (a *analyzer) instAtomicRMW(inst *ir.InstAtomicRMW, f *ir.Func, stack []meta) ir.Instruction {
	values := wrapValues{
		ordering: fromAtomicOrdering(inst.Ordering),
		atomic:   true,
	}
	in := &wrapInstAtomicRMW{InstAtomicRMW: inst, wrapInst: newWrap(inst, values, f, stack, a.count)}
	a.count++
	a.mod.addInst(a.count, in)
	return in
}

func analyze(mod *wrapModule) VisitCallback {
	a := newAnalyzer(mod)

	return func(i ir.Instruction, f *ir.Func, stack []meta) ir.Instruction {
		switch inst := i.(type) {
		case *ir.InstAlloca:
			a.isAlloca[inst] = true

		case *ir.InstCall:
			a.instCall(inst)

		case *ir.InstLoad:
			return a.instLoad(i, inst, f, stack)

		case *ir.InstStore:
			return a.instStore(i, inst, f, stack)

		case *ir.InstFence:
			return a.instFence(inst, f, stack)

		case *ir.InstCmpXchg:
			return a.instCmpXchg(inst, f, stack)
		case *ir.InstAtomicRMW:
			return a.instAtomicRMW(inst, f, stack)

		default:
		}
		return nil
	}
}

func (m *wrapModule) bitseq(sel core.Selection, after bool) core.Bitseq {
	modes := true
	if sel == core.SelectionLoads || sel == core.SelectionStores || sel == core.SelectionPlain {
		modes = false
	}
	return m.get(sel, after).Bitseq(modes, after)
}

func (m *wrapModule) count(atype core.Selection, after bool) int {
	return len(m.get(atype, after))
}

func countBarrier(bc *barrierCount, o core.Ordering) {
	switch o {
	case core.Relaxed:
		bc.Relaxed++
	case core.Acquire:
		bc.Acquire++
	case core.Release:
		bc.Release++
	case core.SeqCst:
		bc.SeqCst++
	default:
	}
}

func (m *wrapModule) barrierCount(sel core.Selection, after bool) barrierCount {
	bc := new(barrierCount)
	for _, i := range m.get(sel, after) {
		countBarrier(bc, i.getOrdering(after))
	}
	return *bc
}
