// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"vsync/logger"
)

func setEnv(env, val string) {
	if err := os.Setenv(env, val); err != nil {
		logger.Fatalf("could not set %v: %v", env, err)
	}
	logger.Printf("%v = %v\n", env, val)
}

var testsDir = "../../tests"

type fileTest struct {
	dir string
	fn  string
}

func (ft fileTest) path() string {
	return filepath.Join(ft.dir, ft.fn)
}

var (
	testCFiles  []fileTest
	testLLFiles []fileTest
)

// load C file examples
func loadFiles() {
	if len(testCFiles) != 0 {
		return
	}
	if val := os.Getenv("TESTS_DIR"); val != "" { //permit:os.Getenv
		testsDir = val
	}

	files, err := ioutil.ReadDir(testsDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() && reIsC.MatchString(file.Name()) {
			testCFiles = append(testCFiles,
				fileTest{
					fn:  file.Name(),
					dir: testsDir,
				})
		}
		if !file.IsDir() && reIsLL.MatchString(file.Name()) {
			testLLFiles = append(testLLFiles,
				fileTest{
					fn:  file.Name(),
					dir: testsDir,
				})
		}
	}
}
