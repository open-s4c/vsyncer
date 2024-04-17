// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"fmt"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/metadata"

	"vsync/logger"
)

var verboseVisitor = true

// DebugVisitor turns on several debugging prints in the visitor.
func DebugVisitor() {
	verboseVisitor = true
}

type meta interface {
	MDAttachments() []*metadata.Attachment
}

// VisitCallback is called in every relevant intruction when visiting the module.
type VisitCallback func(inst ir.Instruction, f *ir.Func, stack []meta) ir.Instruction

func (w *wrapModule) Visit(fun []string, cb VisitCallback) error {
	return visitModule(w.Module, fun, cb)
}

func visitModule(m *ir.Module, fun []string, cb VisitCallback) error {
	for _, foo := range fun {
		for _, f := range m.Funcs {
			if f.Ident() != fmt.Sprintf("@%s", foo) {
				continue
			}
			v := &visitor{visited: make(map[ir.Instruction]bool)}
			v.log("====================== START VISIT ==========================")
			if err := v.visit(f, []meta{f}, cb); err != nil {
				return err
			}
		}
	}
	return nil
}

type visitor struct {
	dep     string
	visited map[ir.Instruction]bool
}

func (v *visitor) enter() {
	v.dep += " "
}

func (v *visitor) leave() {
	v.dep = v.dep[:len(v.dep)-1]
}

func (v *visitor) log(args ...any) {
	if !verboseVisitor {
		return
	}
	logger.Print(v.dep)
	logger.Println(args...)
}

var fstr = fmt.Sprintf

func (v *visitor) logf(str string, args ...any) {
	if !verboseVisitor {
		return
	}
	logger.Print(v.dep)
	logger.Print(fstr(str, args...))
}

func (v *visitor) visitCallee(inst *ir.InstCall, f *ir.Func, stack []meta, cb VisitCallback) error {
	callee := inst.Callee.Ident()
	var err error
	if strings.Contains(callee, "pthread_create") {
		threadRun := inst.Operands()[3]
		ff, ok := (*threadRun).(*ir.Func)
		if !ok {
			if arg, ok := (*threadRun).(*ir.Arg); !ok {
				return fmt.Errorf("could not cast %v", *threadRun)
			} else if ff, ok = arg.Value.(*ir.Func); !ok {
				logger.Warnf("Ignoring function pointer in pthread_create.")
				return nil
			}
		}
		v.enter()
		err = v.visit(ff, append(stack, inst), cb)
		v.leave()
	} else if strings.Contains(callee, "__VERIFIER_thread_create") {
		var ff *ir.Func
		// signature
		//   __VERIFIER_thread_create(attribute, start_routine, arg) --> object
		v.logf("__VERIFIER_thread_create Operands: %v\n", inst.Operands())
		threadRun := inst.Operands()[2]
		v.logf("Operands()[2]: %#v\n", *threadRun)

		if arg, ok := (*threadRun).(*ir.Arg); !ok {
			return fmt.Errorf("could not cast ir.Arg %v", *threadRun)
		} else if ff, ok = arg.Value.(*ir.Func); !ok {
			logger.Warnf("Ignoring function pointer in pthread_create.")
			return nil
		}
		v.enter()
		err = v.visit(ff, append(stack, inst), cb)
		v.leave()
	} else if !strings.Contains(callee[1:], "__assert_fail") && !strings.Contains(callee[1:], "llvm.") &&
		!strings.Contains(callee[1:], "_VERIFIER") && !strings.Contains(callee[1:], "pthread_") {

		switch callee := inst.Callee.(type) {
		case *ir.InlineAsm:
		case *ir.InstLoad:
			// this is a function pointer, ignore this case
			logger.Warnf("%s: ignoring function pointer '%v'", f.Ident(), callee)
		case *ir.Func:
			v.enter()
			err = v.visit(callee, append(stack, inst), cb)
			v.leave()
		case *constant.ExprBitCast:
			if foo, ok := callee.From.(*ir.Func); ok {
				v.enter()
				err = v.visit(foo, append(stack, inst), cb)
				v.leave()
			}
		default:
			panic(fmt.Errorf("cannot convert %T: %v", inst.Callee, inst.Callee))
		}
	}
	return err
}

func (v *visitor) visitInst(inst ir.Instruction, f *ir.Func,
	stack []meta, cb VisitCallback) (bool, ir.Instruction, error) {
	if v.visited[inst] {
		v.log("SKIP: ", inst)
		return true, nil, nil
	}
	v.visited[inst] = true
	v.log("Inst: ", inst)

	var ni ir.Instruction
	switch inst := inst.(type) {
	case *ir.InstAlloca:
		ni = cb(inst, f, append(stack, inst))
	case *ir.InstAtomicRMW:
		ni = cb(inst, f, append(stack, inst))
	case *ir.InstFence:
		ni = cb(inst, f, append(stack, inst))
	case *ir.InstLoad:
		ni = cb(inst, f, append(stack, inst))
	case *ir.InstStore:
		ni = cb(inst, f, append(stack, inst))
	case *ir.InstCmpXchg:
		ni = cb(inst, f, append(stack, inst))
	case *ir.InstCall:
		v.log("visit? ", inst.Callee.Ident())
		v.enter()
		ni = cb(inst, f, append(stack, inst))
		v.leave()
		if err := v.visitCallee(inst, f, stack, cb); err != nil {
			return false, nil, err
		}
	default:
	}

	return ni == nil, ni, nil
}

func (v *visitor) visit(f *ir.Func, stack []meta, cb VisitCallback) error {
	for _, block := range f.Blocks {
		for i, inst := range block.Insts {
			skip, ni, err := v.visitInst(inst, f, stack, cb)
			if err != nil {
				return err
			}
			if !skip {
				block.Insts[i] = ni
			}
		}
	}
	return nil
}
