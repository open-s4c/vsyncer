// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io/ioutil"
	_ "io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"vsync/core"
	"vsync/logger"
	"vsync/module"
	"vsync/tools"
)

func TestMutateCFiles(t *testing.T) {
	loadFiles()
	for _, f := range testCFiles {
		f := f
		fn := f.fn
		t.Run(fmt.Sprintf(filepath.Base(fn)), func(t *testing.T) {
			var (
				args = []string{f.path()}
				fn   = "test.ll"
			)

			err := Compile(fn, args...)
			assert.Nil(t, err)

			m, err := mutate(fn, liftSelection, orderSelection)
			assert.Nil(t, err)

			if m != nil {
				m.Cleanup()
			}
		})
		break
	}
}

const prog = `
int v;

int main() {
	int x = v;
	int y = __atomic_load_n(&v, __ATOMIC_SEQ_CST);
	x++;
	v = x;
	__atomic_store_n(&v, 123, __ATOMIC_SEQ_CST);
	int w = __atomic_exchange_n(&v, 1, __ATOMIC_SEQ_CST);
	return 0;
}
`

func TestMutateAtomic(t *testing.T) {
	// create temporary C file
	f, err := ioutil.TempFile(".", "vsyncer_test.*.c")
	assert.Nil(t, err)
	defer tools.Remove(f.Name())
	_, err = f.WriteString(prog)
	assert.Nil(t, err)
	assert.Nil(t, f.Close())

	// prepare history and filenames
	var (
		args = []string{f.Name()}
		cfg  = module.DefaultConfig()
		fnll = "test.ll"
	)
	logger.SetLevel(logger.DEBUG)

	// compile file to LLVM-IR
	err = Compile(fnll, args...)
	assert.Nil(t, err)
	defer tools.Remove(fnll)

	// analyze input LL file
	mod, err := module.Load(fnll, cfg)
	assert.Nil(t, err)

	// check expected bitseq
	a := mod.Assignment(core.SelectionAtomic)
	bs := a.Bs
	bexp := core.MustFromString("0x3f").Fit(bs.Length())
	assert.True(t, bs.Equals(bexp))

	// mutate to 0x3f
	bitseqFlags[core.SelectionAtomic].value = "0x20"
	mod, err = mutate(fnll, []core.Selection{core.SelectionAtomic})
	assert.Nil(t, err)

	fn2 := "test2.ll"
	err = tools.Dump(mod, fn2)
	assert.Nil(t, err)

	// check new bitseq
	mod2, err := module.Load(fn2, cfg)
	assert.Nil(t, err)
	a = mod2.Assignment(core.SelectionAtomic)
	bs = a.Bs
	assert.True(t, bs.Length() > 0)
	bexp = core.MustFromString("0x20").Fit(bs.Length())
	assert.True(t, bs.Equals(bexp))

	// cleanup
	mod.Cleanup()
	mod2.Cleanup()
}
