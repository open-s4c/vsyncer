// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"sync"

	"github.com/llir/llvm/asm"
	"github.com/llir/llvm/ir"

	"vsync/core"
	"vsync/logger"
)

type wrapModule struct {
	*ir.Module
	sync.Mutex
	imap wrapInstSelection
}

func loadModule(fn string, cfg Config) (*wrapModule, error) {
	logger.Infof("Parse '%s'", fn)

	mod, err := asm.ParseFile(fn)
	if err != nil {
		return nil, err
	}
	wmod := newWrapModule(mod)

	logger.Infof("Analyze '%s'", fn)
	if err := wmod.Visit(cfg.EntryFunc, analyze(wmod), cfg); err != nil {
		return nil, err
	}
	return wmod, nil
}

func newWrapModule(mod *ir.Module) *wrapModule {
	imap := make(wrapInstSelection)

	return &wrapModule{
		Module: mod,
		imap:   imap,
	}
}

func (wm *wrapModule) addInst(id int, in wrapInstruction) {
	_, has := wm.imap[id]
	if has {
		logger.Fatalf("error: already has instruction '%v'", id)
	}
	wm.imap[id] = in
}

func (wm *wrapModule) get(sel core.Selection, after bool) wrapInstSelection {
	insts := make(wrapInstSelection)
	for ir, i := range wm.imap {
		if matchInstSelection(i, sel.Group(), after) {
			insts[ir] = i
		}
	}
	return insts
}

func getInstSelection(i wrapInstruction, after bool) core.Selection {
	switch i := i.(type) {
	case *wrapInstAtomicRMW:
		return core.SelectionRMWs
	case *wrapInstCmpXchg:
		return core.SelectionRMWs
	case *wrapInstFence:
		return core.SelectionFences
	case *wrapInstLoad:
		if after {
			if i.after.atomic {
				return core.SelectionAtomicLoads
			}
			return core.SelectionPlainLoads
		}
		if i.before.atomic {
			return core.SelectionAtomicLoads
		}
		return core.SelectionPlainLoads
	case *wrapInstStore:
		if after {
			if i.after.atomic {
				return core.SelectionAtomicStores
			}
			return core.SelectionPlainStores
		}
		if i.before.atomic {
			return core.SelectionAtomicStores
		}
		return core.SelectionPlainStores
	default:
		logger.Printf("invalid type: %T", i)
	}
	return core.SelectionInvalid
}

func matchInstSelection(i wrapInstruction, sel []core.Selection, after bool) bool {
	is := getInstSelection(i, after)
	for _, s := range sel {
		if is == s {
			return true
		}
	}
	return false
}
