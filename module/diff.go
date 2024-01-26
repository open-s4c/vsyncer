// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/metadata"

	"vsync/core"
	"vsync/logger"
)

var (
	rlxColor = color.New(color.FgRed).SprintFunc()
	relColor = color.New(color.FgGreen).SprintFunc()
	acqColor = color.New(color.FgYellow).SprintFunc()
	seqColor = color.New(color.FgCyan).SprintFunc()
	naColor  = color.New(color.FgBlue).SprintFunc()

	changeColor = color.New(color.BgCyan, color.FgBlack).SprintFunc()
)

const suffixLength = 4

func (m *wrapModule) Diff() []*diffEntry {
	var diff []*diffEntry
	for _, id := range m.imap.sortedKeys() {
		in := m.imap[id]
		if entry := in.diff(); entry != nil {
			diff = append(diff, entry)
		}
	}
	return diff
}

type diffEntry struct {
	Loc            Loc
	AtomicBefore   bool
	AtomicAfter    bool
	OrderingBefore core.Ordering
	OrderingAfter  core.Ordering
	Delete         bool
	FuncName       string
	CloneName      string
	Name           string
}

var reClone = regexp.MustCompile(`(.*)__vsyncer_expand_[0-9]+$`)

func (inst *wrapInst) diff() *diffEntry {

	switch {
	case inst.before.atomic != inst.after.atomic:
		return &diffEntry{
			Loc:          getLoc(inst.stack),
			AtomicBefore: inst.before.atomic,
			AtomicAfter:  inst.after.atomic,
			Name:         fmt.Sprintf("%T", inst.inst),
		}
	case !inst.after.atomic:
		return nil
	case inst.before.ordering != inst.after.ordering:
		entry := diffEntry{
			Loc:            getLoc(inst.stack),
			AtomicBefore:   true,
			AtomicAfter:    true,
			OrderingBefore: inst.before.ordering,
			OrderingAfter:  inst.after.ordering,
			Name:           fmt.Sprintf("%T", inst.inst),
		}

		if _, ok := inst.inst.(*ir.InstFence); ok && inst.after.ordering == core.Relaxed {
			entry.Delete = true
		}

		entry.FuncName = reClone.ReplaceAllString(inst.f.GlobalName, "${1}")
		if entry.FuncName != inst.f.GlobalName {
			entry.CloneName = inst.f.GlobalName
		}

		return &entry
	default:
		return nil
	}
}

func getLoc(stack []meta) Loc {
	for k := len(stack) - 1; k >= 0; k-- {
		if loc := readLoc(stack[k]); !strings.Contains(loc.Filename, "vsync/atomic") {
			return loc
		}
	}
	return Loc{}
}

// Loc represents a code location extracted from the IR.
type Loc struct {
	Filename  string
	Directory string
	Line      int64
	Column    int64
}

func (loc *Loc) update(line, col int64, filename string, directory string) bool {
	// update location
	if loc.Line == 0 {
		loc.Line = line
	}
	if loc.Column == 0 {
		loc.Column = col
	}
	if loc.Filename == "" {
		loc.Filename = filename
	}
	if loc.Directory == "" {
		loc.Directory = directory
	}

	// return if enough info availabl
	return loc.Filename != "" && loc.Line != 0
}

func readLoc(md meta) Loc {
	var (
		loc  Loc
		node interface{}
		done bool
	)
	for _, ma := range md.MDAttachments() {
		if ma.Name == "dbg" {
			node = ma.Node
			break
		}
	}
	for node != nil && !done {
		var (
			line      int64
			col       int64
			filename  string
			directory string
		)
		switch n := node.(type) {
		case *metadata.DILocation:
			line = n.Line
			col = n.Column
			node = n.Scope
		case *metadata.DILexicalBlock:
			filename = n.File.Directory + "/" + n.File.Filename
			directory = n.File.Directory
			line = n.Line
			col = n.Column
			node = n.Scope
		case *metadata.DISubprogram:
			filename = n.File.Directory + "/" + n.File.Filename
			directory = n.File.Directory
			line = n.Line
			node = n.Scope
		case *metadata.DIFile:
			filename = n.Directory + "/" + n.Filename
			directory = n.Directory
			node = nil
		default:
		}
		done = loc.update(line, col, filename, directory)
	}
	return loc
}

func printDiffEntry(i int, d *diffEntry) error {
	var (
		line string
		err  error
	)

	if line, err = readLineFromFile(d.Loc.Filename, d.Loc.Line); err != nil {
		return err
	}
	line = annotate(line, d.Loc.Column)

	to, err := locChangeVsync(d)
	if err != nil {
		return err
	}

	logger.Printf("[%d] %s:%d:%d\n", i, d.Loc.Filename, d.Loc.Line, d.Loc.Column)
	logger.Printf("%s %s\n\n", line, to)
	return nil
}

func readLineFromFile(fn string, line int64) (string, error) {
	f, err := os.Open(fn)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Warnf("error closing file: %v", err)
		}
	}()
	scanner := bufio.NewScanner(f)
	for i := int64(0); scanner.Scan(); i++ {
		if i+1 == line {
			return scanner.Text(), scanner.Err()
		}
	}

	return "", fmt.Errorf("could not find line %d in file %s", line, fn)
}

func annotate(text string, col int64) string {
	if text == "" {
		return text
	}
	str := ""
	for i := int64(0); i < col-1; i++ {
		if text[i] == '\t' {
			str += "\t"
		} else {
			str += " "
		}
	}
	str += "^~~~~~~~~~ "
	return text + "\n" + str
}

func colSprintf(o core.Ordering) func(a ...interface{}) string {
	switch o {
	case core.Relaxed:
		return rlxColor
	case core.Acquire:
		return acqColor
	case core.Release:
		return relColor
	case core.SeqCst:
		return seqColor
	default:
		logger.Fatalf("not color implemented: %v", o)
	}
	return nil
}

func withColor(o core.Ordering) string {
	return colSprintf(o)(o)
}

func orderSuffix(o core.Ordering) string {
	sprintf := colSprintf(o)
	switch o {
	case core.Relaxed:
		return sprintf("_rlx")
	case core.Acquire:
		return sprintf("_acq")
	case core.Release:
		return sprintf("_rel")
	case core.SeqCst:
		return ""
	default:
		return ""
	}
}

func locChangeVsync(d *diffEntry) (string, error) {
	if d.AtomicAfter && d.AtomicBefore {
		return changeOrdering(d)
	}
	if d.AtomicAfter {
		return seqColor(fmt.Sprintf("change %s to atomic", d.Name)), nil
	}
	return rlxColor(fmt.Sprintf("change %s to non-atomic", d.Name)), nil
}

func changeOrdering(d *diffEntry) (string, error) {
	if d.Delete {
		return naColor("remove it"), nil
	}

	if !strings.Contains(d.FuncName, "vatomic") {
		return fmt.Sprintf("change %s to %s", d.Name, withColor(d.OrderingAfter)), nil
	}
	to := d.FuncName

	if strings.HasSuffix(to, "_rel") ||
		strings.HasSuffix(to, "_rlx") ||
		strings.HasSuffix(to, "_acq") {
		to = to[:len(to)-suffixLength]
	}

	return seqColor(to) + orderSuffix(d.OrderingAfter), nil
}
