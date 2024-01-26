// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"fmt"

	"vsync/core"
	"vsync/logger"
	"vsync/tools"
)

// PrintSummary displays at standard output a summary of the module and its recorded mutations.
func (h *History) PrintSummary() {
	logger.Println("== SUMMARY ===================================")
	logger.Println()
	logger.Println("File")
	logger.Printf("  %s\n", h.strFiles(0))
	logger.Println()
	logger.Println("Operations")
	var (
		plain  = h.countDiff(core.SelectionPlainLoads, 0)
		atomic = h.countDiff(core.SelectionAtomicLoads, 0)
	)
	logger.Printf("  Plain  loads  : %v\n", plain)
	logger.Printf("  Atomic loads  : %v\n", atomic)
	plain = h.countDiff(core.SelectionPlainStores, 0)
	atomic = h.countDiff(core.SelectionAtomicStores, 0)
	logger.Printf("  Plain  stores : %v\n", plain)
	logger.Printf("  Atomic stores : %v\n", atomic)

	var (
		rmws   = h.countDiff(core.SelectionRMWs, 0)
		fences = h.countDiff(core.SelectionFences, 0)
	)

	logger.Printf("  RMWs          : %v\n", rmws)
	logger.Printf("  Fences        : %v\n", fences)
	logger.Println()

	logger.Println("Memory ordering")
	logger.Printf("  SeqCst  : %s\n", h.barrierCountDiff(core.SeqCst, 0))
	logger.Printf("  Release : %s\n", h.barrierCountDiff(core.Release, 0))
	logger.Printf("  Acquire : %s\n", h.barrierCountDiff(core.Acquire, 0))
	logger.Printf("  Relaxed : %s\n", h.barrierCountDiff(core.Relaxed, 0))
	logger.Println()

	logger.Println("Assignments")
	x := []struct {
		text  string
		atype core.Selection
	}{
		{"[L] Loads  ", core.SelectionLoads},
		{"[S] Stores ", core.SelectionStores},
		{"[A] Atomics", core.SelectionAtomic},
		{"[F] Fences ", core.SelectionFences},
		{"[X] RMWs   ", core.SelectionRMWs},
	}
	for _, e := range x {
		logger.Printf("  %s : %v\n", e.text, h.bitseqDiff(e.atype, 0))
	}
	logger.Println()

}

// PrintDiff displays the source code difference between the module's initial state and final mutation.
func (h *History) PrintDiff() error {
	if h.length() <= 1 {
		return nil
	}

	logger.Println("== CODE DIFF =================================")
	logger.Println()

	// reload original file and reapply mutations
	fn := h.hist[0]
	first, err := loadModule(fn, h.cfg)
	if err != nil {
		return err
	}
	logger.Debugf("Initial assignment: %v", first.Assignment(core.SelectionAtomic))
	for _, a := range h.mutations.recorded {
		logger.Debugf("Target assignment: %v", a)
		if err := first.mutate(a.Bs, a.Sel); err != nil {
			return err
		}
	}
	for _, a := range h.mutations.current {
		logger.Debugf("Target assignment: %v", a)
		if err := first.mutate(a.Bs, a.Sel); err != nil {
			return err
		}
	}
	logger.Debugf("Final assignment: %v", first.Assignment(core.SelectionAtomic))
	err = tools.Dump(first, "something.ll")
	if err != nil {
		return err
	}
	// print difference
	diff := first.Diff()
	for i, d := range diff {
		if err := printDiffEntry(i, d); err != nil {
			return fmt.Errorf("cannot print diff: %v", err)
		}
	}

	return nil
}

func (h *History) countDiff(sel core.Selection, i int) string {
	if i >= len(h.hist) {
		return ""
	}
	last := i == len(h.hist)-1

	sep := " \t--> "
	if i == 0 {
		sep = ""
	}

	before := h.mods[i].count(sel, false)
	after := h.mods[i].count(sel, true)
	cur := fmt.Sprintf("%v", before)
	if last {
		cur = fmt.Sprintf("%v", after)
	}
	if before != after {
		cur = changeColor(cur)
	}

	return fmt.Sprintf("%s%s%s", sep, cur, h.countDiff(sel, i+1))
}

func (h *History) barrierCountDiff(ordering core.Ordering, i int) string {
	if i >= len(h.hist) {
		return ""
	}
	last := i == len(h.hist)-1
	getVal := func(m *wrapModule) string {
		bcBefore := m.barrierCount(core.SelectionAtomic, false)
		bcAfter := m.barrierCount(core.SelectionAtomic, true)
		bc := bcBefore
		if last {
			bc = bcAfter
		}
		var txt string
		switch ordering {
		case core.SeqCst:
			txt = fmt.Sprintf("%v", bc.SeqCst)
			if bcBefore.SeqCst != bcAfter.SeqCst {
				txt = changeColor(txt)
			}
		case core.Acquire:
			txt = fmt.Sprintf("%v", bc.Acquire)
			if bcBefore.Acquire != bcAfter.Acquire {
				txt = changeColor(txt)
			}
		case core.Release:
			txt = fmt.Sprintf("%v", bc.Release)
			if bcBefore.Release != bcAfter.Release {
				txt = changeColor(txt)
			}
		case core.Relaxed:
			txt = fmt.Sprintf("%v", bc.Relaxed)
			if bcBefore.Relaxed != bcAfter.Relaxed {
				txt = changeColor(txt)
			}
		default:
		}
		return txt
	}

	sep := " \t--> "
	if i == 0 {
		sep = ""
	}
	return fmt.Sprintf("%s%v%s", sep, getVal(h.mods[i]), h.barrierCountDiff(ordering, i+1))
}

func (h *History) bitseqDiff(sel core.Selection, i int) string {
	if i >= len(h.hist) {
		return ""
	}
	last := i == len(h.hist)-1

	sep := " --> "
	if i == 0 {
		sep = ""
	}

	before := h.mods[i].bitseq(sel, false)
	after := h.mods[i].bitseq(sel, true)
	cur := fmt.Sprintf("%v", before)

	if last {
		cur = fmt.Sprintf("%v", after)
	}

	if !before.Equals(after) {
		cur = changeColor(cur)
	}

	return fmt.Sprintf("%s%v%s", sep, cur, h.bitseqDiff(sel, i+1))
}
