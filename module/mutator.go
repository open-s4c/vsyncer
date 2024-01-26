// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"fmt"

	"vsync/core"
	"vsync/logger"
)

const u2 = 2

// Mutate transforms the LLVM instructions of the module according to an assignment.
func (m *History) Mutate(a core.Assignment) error {
	m.appendMutation(a)
	return m.wrapModule.mutate(a.Bs, a.Sel)
}

func (m *wrapModule) mutate(bs core.Bitseq, sel core.Selection) error {
	m.Lock()
	defer m.Unlock()

	wi := m.get(sel, true)
	keys := wi.sortedKeys()

	// assert length of wi and bs match
	// iterate sorted, apply mutation
	var err error
	if sel.Binary() {
		err = bs.Translate(u2, func(k int, val int) error {
			in := wi.get(keys[k])
			o := mapOrdering(in, val)
			if o == core.Invalid {
				return fmt.Errorf("bitseq with an invalid ordering for operation: %v", mapInstruction(in))
			}
			if !in.isAtomic(true) {
				logger.Fatal("instruction is not atomic")
			}
			in.setOrdering(o)
			return nil
		})
	} else {
		err = bs.Translate(1, func(k int, val int) error {
			in := wi.get(keys[k])
			switch val {
			case 1:
				in.setAtomic(true)
				in.setOrdering(core.SeqCst)
			case 0:
				in.setAtomic(false)
				in.setOrdering(core.Invalid)
			default:
				logger.Fatal("unexpected value:", val)
			}
			return nil
		})
	}
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	return nil
}
