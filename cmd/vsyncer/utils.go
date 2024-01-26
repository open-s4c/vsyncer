// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"

	"vsync/logger"
)

// IsArgsn ensures there are 1 or more arguments
func IsArgsn(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no input file specified")
	}
	return nil
}

var (
	reIsC     = regexp.MustCompile(`.*\.c$`)
	reIsCPP   = regexp.MustCompile(`.*\.cpp$`)
	reIsLL    = regexp.MustCompile(`(.*)\.ll$`)
	reIsLLOrC = regexp.MustCompile(`(.*)\.(ll|c|C|cpp|cxx)$`)
)

func onlyLL(args ...string) bool {
	if len(args) == 0 {
		return false
	}
	for _, a := range args {
		if !reIsLL.MatchString(a) {
			return false
		}
	}
	return true
}

func baseLL(fn string) string {
	return reIsLL.ReplaceAllString(fn, "${1}")
}

func base(fn string) string {
	return reIsLLOrC.ReplaceAllString(fn, "${1}")
}

func hasToCompile(args []string) bool {
	for _, a := range args {
		if !reIsLL.MatchString(a) {
			return true
		}
	}
	return false
}

type fnGen func(string) string

// newOutputGenerator returns a filename generator function based on the command line arguments.
func newOutputGenerator(args []string) fnGen {
	return func(suffix string) string {
		b := findBase(args)
		fn := fmt.Sprintf("%s%s.ll", b, suffix)
		logger.Debugf("[output gen] fn = %s", fn)
		return fn
	}
}

func findBase(args []string) string {
	var a []string
	a = append(a, args...)
	return base(findFile(a))
}

func findFile(args []string) string {
	for _, a := range args {
		if reIsLLOrC.MatchString(a) {
			return a
		}
	}
	return ""
}
