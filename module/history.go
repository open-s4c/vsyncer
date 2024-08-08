// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"fmt"
	"regexp"

	"github.com/llir/llvm/asm"

	"vsync/core"
	"vsync/logger"
	"vsync/tools"
)

// History represents the mutation history of an LLVM IR module.
type History struct {
	*wrapModule
	cfg Config

	base  string
	count int

	hist []string
	mods []*wrapModule

	mutations struct {
		current  []core.Assignment
		recorded []core.Assignment
	}
}

// Load parses and analyzes an LLVM IR module.
func Load(fn string, cfg Config) (*History, error) {
	var efn = fn

	if cfg.Expand {
		logger.Infof("Parse '%s'", fn)
		mod, err := asm.ParseFile(fn)
		if err != nil {
			return nil, err
		}

		logger.Infof("Expand '%s'", fn)
		if err := visitModule(mod, cfg.EntryFunc, expandVisitor(mod), cfg); err != nil {
			return nil, err
		}

		efn = genName(fn, ".expand")
		if err := tools.Dump(mod, efn); err != nil {
			return nil, err
		}
		cfg.Expand = false
	}

	// load wrapped module
	wmod, err := loadModule(efn, cfg)
	if err != nil {
		return nil, err
	}

	return &History{
		wrapModule: wmod,
		cfg:        cfg,
		count:      0,
		base:       fn,
		hist:       []string{efn},
		mods:       []*wrapModule{wmod},
	}, nil
}

// Record saves the current mutation extending the history of the module.
func (h *History) Record() error {
	fn := genName(h.base, h.new())
	if err := tools.Dump(h.Module, fn); err != nil {
		return err
	}
	wm, err := loadModule(fn, h.cfg)
	if err != nil {
		return err
	}
	h.add(fn, wm)
	h.wrapModule = wm

	h.recordMut()
	return nil
}

// Forget drops non-recorded mutations
func (h *History) Forget() error {
	fn := h.last()
	wm, err := loadModule(fn, h.cfg)
	if err != nil {
		return err
	}
	h.mods[len(h.mods)-1] = wm
	h.wrapModule = wm
	h.clearMut()
	return nil
}

// Cleanup removes temporary files.
func (h *History) Cleanup() {
	for _, fn := range h.hist {
		logger.Debugf("Removing history file '%s'", fn)
		if err := tools.Remove(fn); err != nil {
			logger.Debug(err)
		}
	}
}

func (h *History) recordMut() {
	muts := h.mutations.current
	if len(muts) > 0 {
		h.mutations.recorded = append(h.mutations.recorded, muts...)
	}
	h.clearMut()
}

func (h *History) clearMut() {
	h.mutations.current = nil
}

func (h *History) appendMutation(mut core.Assignment) {
	h.mutations.current = append(h.mutations.current, mut)
}

func (h *History) add(name string, mod *wrapModule) {
	h.hist = append(h.hist, name)
	h.mods = append(h.mods, mod)
}

func (h *History) last() string {
	return h.hist[len(h.hist)-1]
}

func (h *History) strFiles(i int) string {
	if i >= len(h.hist) {
		return ""
	}
	sep := " --> "
	if i == 0 {
		sep = ""
	}
	return fmt.Sprintf("%s%s%s", sep, h.hist[i], h.strFiles(i+1))
}

func (h *History) length() int {
	return len(h.hist)
}

func (h *History) new() string {
	h.count++
	return fmt.Sprintf("_%d", h.count)
}

var reLL = regexp.MustCompile(`(.*)\.ll$`)

type fnGen func(string) string

func genName(fn, suffix string) string {
	b := reLL.ReplaceAllString(fn, "${1}")
	return fmt.Sprintf("%s%s.ll", b, suffix)
}
