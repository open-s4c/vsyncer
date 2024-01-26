// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"vsync/tools"
)

func TestInfoCFiles(t *testing.T) {
	loadFiles()
	for _, f := range testCFiles {
		fn := f.fn
		t.Run(fmt.Sprintf(filepath.Base(fn)), func(t *testing.T) {
			output := fn + ".out.ll"
			err := Compile(output, f.path())
			assert.Nil(t, err)
			err = Info(output, []string{f.path()})
			assert.Nil(t, err)
			defer tools.Remove(output)

		})
		t.Run(fmt.Sprintf("%s+compile", filepath.Base(fn)), func(t *testing.T) {
			output := fn + ".out.ll"
			err := Info(output, []string{f.path()})
			assert.Nil(t, err)
			defer tools.Remove(output)

		})
	}
}

func TestInfoLLFiles(t *testing.T) {
	for _, f := range testLLFiles {

		fn := f.fn
		t.Run(fmt.Sprintf(filepath.Base(fn)), func(t *testing.T) {
			var (
				args      = []string{f.path()}
				outputGen = newOutputGenerator(args)
				output    = outputGen("")
			)
			err := Info(output, args)
			assert.Nil(t, err)
			defer tools.Remove(output)

		})
	}
}
