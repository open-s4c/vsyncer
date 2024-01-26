// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"vsync/logger"
	"vsync/module"
	"vsync/tools"
)

func init() {
	var infoCmd = cobra.Command{
		Use:   "info <input.ll|input.c>",
		Short: "Prints information about in the input file(s).",
		Args:  IsArgsn,

		DisableFlagsInUseLine: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				outputGen = newOutputGenerator(args)
				fn        = outputGen("")
			)

			return Info(fn, args)
		},
	}

	rootCmd.AddCommand(&infoCmd)
}

// Info compiles input, analyzes result, and prints summary.
func Info(fn string, args []string) error {
	if hasToCompile(args) {
		if err := Compile(fn, args...); err != nil {
			return err
		}
		defer tools.Remove(fn)
	}
	logger.Debugf("Info %s", fn)

	m, err := module.Load(fn, moduleConfig())
	if err != nil {
		return verror(internalError, err)
	}
	defer m.Cleanup()
	m.PrintSummary()
	return nil
}

func moduleConfig() module.Config {
	return module.Config{
		EntryFunc: rootFlags.entryFunc,
		Expand:    rootFlags.expand,
	}
}
