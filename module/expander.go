// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"fmt"
	"log"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/metadata"

	"vsync/logger"
)

func expandVisitor(mod *ir.Module) VisitCallback {
	// count number of clones of a function
	count := make(map[string]int)

	return func(inst ir.Instruction, _ *ir.Func, stack []meta) ir.Instruction {
		switch inst := inst.(type) {
		case *ir.InstCall:
			// calle function name
			fname := inst.Callee.Ident()[1:]

			if strings.Contains(fname, "llvm") {
				return nil
			}

			// ignore already expanded functions
			if strings.Contains(fname, "__vsyncer_expand_") {
				logger.Debugf("skip %s", fname)
				return nil
			}

			// only expand functions containing "vatomic"
			if !strings.Contains(fname, "vatomic") {
				return nil
			}

			// increment clone count, but only clone after 1
			count[fname]++

			// clone call instruction
			in := new(ir.InstCall)
			if err := copier.Copy(in, inst); err != nil {
				log.Fatalln(err)
			}

			// clone function
			cloneFname := fmt.Sprintf("%s__vsyncer_expand_%d", fname, count[fname]-1)
			in.Callee = cloneFunc(mod, fname, cloneFname)

			// return call instruction to replace current one
			if verboseVisitor {
				logger.Debugf("clonedFunc: %v", in.Callee.Ident())
				logger.Debugf("clonedCall: %v", in.LLString())
			}
			return in
		default:
		}
		return nil
	}
}

func cloneFunc(m *ir.Module, fname, cloneFname string) *ir.Func {
	var f *ir.Func

	// find target func
	for _, fun := range m.Funcs {
		if fun.GlobalIdent.GlobalName == fname {
			f = fun
			break
		}
	}

	if f == nil {
		log.Fatalf("could not find function: %v", fname)
	}

	// clone function
	cloneFunc := new(ir.Func)
	if err := copier.Copy(cloneFunc, f); err != nil {
		log.Fatalln(err)
	}

	// set new name and reset function ID
	cloneFunc.SetName(cloneFname)
	cloneFunc.SetID(-1)

	// add function to module
	m.Funcs = append(m.Funcs, cloneFunc)

	// clone function metadata, duplicating !dbg symbols
	cloneFunc.Metadata = nil
	for _, ma := range f.Metadata.MDAttachments() {
		if ma.Name != "dbg" {
			cloneFunc.Metadata = append(cloneFunc.Metadata, ma)
			continue
		}
		sp, ok := ma.Node.(*metadata.DISubprogram)
		if !ok {
			continue
		}
		newSp := new(metadata.DISubprogram)
		if err := copier.Copy(newSp, sp); err != nil {
			logger.Fatal("could not clone debug metadata", err)
		}
		// reset metadata ID
		newSp.MetadataID.SetID(-1)

		cloneFunc.Metadata = append(cloneFunc.Metadata,
			&metadata.Attachment{
				Name: fname, // use original function name
				Node: newSp,
			})
		// add metadata to module
		m.MetadataDefs = append(m.MetadataDefs, newSp)
	}

	return cloneFunc
}
