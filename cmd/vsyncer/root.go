// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

// Package main is the main vsyncer program of this project.
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"vsync/checker"
	"vsync/logger"
	"vsync/tools"
)

var rootCmd = cobra.Command{
	Use:           "vsyncer",
	Short:         "",
	Long:          "",
	SilenceUsage:  true,
	SilenceErrors: true,

	TraverseChildren: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run 'vsyncer -h' for help")
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch rootFlags.log {
		case "INFO":
			logger.SetLevel(logger.INFO)
		case "WARN":
			logger.SetLevel(logger.WARN)
		default:
			logger.SetLevel(logger.ERROR)
		}
		if rootFlags.debug {
			logger.SetLevel(logger.DEBUG)
		}
		if rootFlags.quiet {
			logger.SetFileDescriptor(nil)
		}
	},
}

func init() {
	tools.RegEnv("VSYNCER_DEFAULT_CHECKER", "genmc", "Default model checker")
	tools.RegEnv("VSYNCER_DEFAULT_ENTRY_FUNC", "main", "Default entry functions for analysis")

	helpMessage :=
		`vsyncer -- Verification and optimization of concurrent code on WMM`

	helpMessage += "\n\nEnvironment Variables:"
	for _, ev := range tools.GetEnvvars() {
		helpMessage += "\n  " + ev.Name + " " +
			"(default: \"" + ev.Defv + "\")\n\t" + ev.Desc
	}
	rootCmd.Long = helpMessage

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&rootFlags.log, "log", "ERROR", "log level (ERROR|INFO|WARN)")
	flags.StringVarP(&rootFlags.checker, "checker", "c", tools.GetEnv("VSYNCER_DEFAULT_CHECKER"), "target checker (genmc|dartagnan|mock)")
	flags.StringVarP(&rootFlags.outputFn, "output", "o", "", "output LLVM file")
	flags.BoolVar(&rootFlags.expand, "expand", true, "expand vatomic functions")
	flags.BoolVarP(&rootFlags.debug, "debug", "d", false, "set debug mode")
	flags.BoolVarP(&rootFlags.quiet, "quiet", "q", false, "do not produce output")
	flags.StringSliceVar(&rootFlags.entryFunc, "entry-func",
		strings.Split(tools.GetEnv("VSYNCER_DEFAULT_ENTRY_FUNC"), ","),
		"list of entry functions")

	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	initOptimize()
}

var reExitStatus = regexp.MustCompile("^exit status [0-9]+$")

func getCheckerID() checker.ID {
	return checker.ParseID(rootFlags.checker)
}

var rootFlags struct {
	log       string
	debug     bool
	outputFn  string
	checker   string
	entryFunc []string
	expand    bool
	quiet     bool

	expandOnly bool
	skipFunc   []string
}

type errCode struct {
	err  error
	code int
}

func handlePanic() {
	e := recover()
	if e == nil {
		return
	}
	exit, ok := e.(errCode)
	if !ok {
		panic(e)
	}
	if exit.err != nil {
		logger.Printf("panic: %v\n", exit.err)
	}
}

func main() {
	if !rootFlags.debug {
		defer handlePanic()
	}
	if err := rootCmd.Execute(); err != nil {
		var (
			code = getErrorCode(err)
			msg  = getErrorMessage(err)
		)

		if match := reExitStatus.MatchString(msg); !match && msg != "" {
			logger.Println(msg)
		}
		os.Exit(code)
	}
}
