// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"errors"

	"github.com/spf13/cobra"

	"vsync/checker"
	"vsync/logger"
	"vsync/tools"
)

const compileDoc = `
Compiles input file(s) to LLVM-IR with clang with required changes to desired
checker. For all checkers, the following options are passed to clang

   -Xclang -disable-O0-optnone -g -S -emit-llvm

Use CFLAGS to pass further compilation flags and set CLANG to select the path
to the clang compiler.
`

func init() {
	var compileCmd = cobra.Command{
		Use:   "compile [flags] <input.c>",
		Short: "Compiles input file with clang",
		Long:  compileDoc,
		Args:  IsArgsn,

		DisableFlagsInUseLine: true,
		TraverseChildren:      true,

		RunE: func(cmd *cobra.Command, args []string) error {
			output := newOutputGenerator(append([]string{rootFlags.outputFn}, args...))
			return Compile(output(""), args...)
		},
	}

	rootCmd.AddCommand(&compileCmd)
}

// Compile takes an output file and arguments from command line
func Compile(output string, args ...string) error {

	switch {
	case len(args) == 0:
		return errors.New("no file argument given")

	// When only one argument given, that can be a .c or an .ll file
	case len(args) == 1 && reIsLL.MatchString(args[0]):
		if err := tools.FileExists(args[0]); err != nil {
			// If the argument is already an .ll file, simply return.
			return err
		}
		return tools.CopyFile(args[0], output)

	// Currently we do support multiple arguments
	case onlyLL(args...):
		logger.Fatal("linking multiple .ll not implemented yet")
	default:
	}

	if err := compileSources(output, args); err != nil {
		return verror(compilerError, err)
	}
	return nil
}

func compileSources(output string, args []string) error {
	for _, f := range args {
		if !reIsC.MatchString(f) && !reIsCPP.MatchString(f) {
			continue
		}
		if err := tools.FileExists(f); err != nil {
			return err
		}
	}

	// The arguments are compilable and exist, so now we do actual compilation.
	getOptions := checker.CompileOptions(getCheckerID())
	return tools.Compile(args, output, getOptions())
}
