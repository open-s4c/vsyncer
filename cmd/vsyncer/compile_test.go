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

func TestCompileOK(t *testing.T) {
	loadFiles()
	for _, f := range testCFiles {
		fn := f.fn
		t.Run(fmt.Sprintf(filepath.Base(fn)), func(t *testing.T) {
			output := fn + ".out.ll"
			err := Compile(output, f.path())
			assert.Nil(t, err)
			defer tools.Remove(output)
		})
	}
}
