// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"vsync/checker"
	"vsync/logger"
	"vsync/module"
	"vsync/tools"
)

var checkFlags = struct {
	opts        []string
	memoryModel string
	csvFile     string
	timeout     time.Duration
}{}

var checkCmd = cobra.Command{
	Use:   "check [flags] <input.ll|input.c>",
	Short: "Checks input file given a mutation bitseq",
	Args:  IsArgsn,
	RunE:  checkRun,

	DisableFlagsInUseLine: true,
}

func init() {
	tools.RegEnv("VSYNCER_DEFAULT_MEMMODEL", "imm", "Default memory model")

	flags := checkCmd.PersistentFlags()
	flags.StringVar(&checkFlags.csvFile, "csv-log", "", "CSV file to append the final result to ")
	flags.DurationVar(&checkFlags.timeout, "timeout", 0, "Check timeout, e.g., 1s for 1 second, 1m for 1 minute.\nCheck will fail if the model checker did not finish within the given time.\ntimeout 0 is equivalent to no timeout")
	addCheckFlags(flags)
	addMutateFlags(flags)
	rootCmd.AddCommand(&checkCmd)
}

func addCheckFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&checkFlags.memoryModel, "memory-model", "m", tools.GetEnv("VSYNCER_DEFAULT_MEMMODEL"), "memory model")
	flags.SetInterspersed(false)
}

func checkResults(result checker.CheckResult, m *module.History, dur time.Duration) (err error) {
	if result.Status != checker.CheckOK {
		logger.Println()
		logger.Println("== OUTPUT ====================================")
		logger.Println()
		logger.Println(result.Output)
		err = vfail(result.Status, fmt.Errorf("%s", err))
	}
	logger.Println()

	if err := m.PrintDiff(); err != nil {
		logger.Debug(err)
	}

	m.PrintSummary()
	logger.Printf("Status\n  %v\n\n", result.Status)
	logger.Printf("Elapsed time\n  %v\n", dur)
	logger.Println()
	return
}

func checkRun(_ *cobra.Command, args []string) (err error) {
	var (
		outputGen = newOutputGenerator(args)
		fn        = outputGen("")
		ts        = time.Now()
		result    checker.CheckResult
		checkerID = getCheckerID()
		mcVersion = ""
		mm        = checker.ParseMemoryModel(checkFlags.memoryModel)
		cxt       = context.Background()
	)
	defer func() {
		csvReport{
			name:          fn,
			checker:       checkerID,
			version:       mcVersion,
			memoryModel:   mm,
			duration:      time.Since(ts),
			status:        result.Status,
			numExecutions: result.NumExecutions,
			err:           err,
		}.save(checkFlags.csvFile)
	}()

	if hasToCompile(args) {
		if err = Compile(fn, args...); err != nil {
			return
		}
		defer tools.Remove(fn)
	}

	var m *module.History
	if m, err = mutate(fn, liftSelection, orderSelection); err != nil {
		return
	}
	if m == nil {
		return errors.New("unexpected nil history pointer")
	}
	defer m.Cleanup()

	var chkr checker.Tool
	if chkr, err = newChecker(checkerID, mm); err != nil {
		return
	}

	if checkFlags.timeout != 0 {
		var cancel context.CancelFunc
		cxt, cancel = context.WithTimeout(cxt, checkFlags.timeout)
		defer cancel()
	}

	result, err = chkr.Check(cxt, m)
	mcVersion = chkr.GetVersion()
	if err != nil {
		logger.Debugf("error in checker: %v\n", err)
		err = verror(checkerError, err)
		return
	}

	err = checkResults(result, m, time.Since(ts))
	if fn := rootFlags.outputFn; fn != "" {
		if lerr := tools.Dump(m, fn); lerr != nil {
			logger.Debug(lerr)
		}
	}
	return
}

func newChecker(cid checker.ID, mm checker.MemoryModel) (checker.Tool, error) {
	if mm == checker.InvalidMemoryModel {
		err := fmt.Errorf("error: invalid memory model '%v'", mm)
		return nil, verror(internalError, err)
	}

	switch cid {
	case checker.GenmcID:
		return checker.NewGenMC(mm, 1), nil
	case checker.DartagnanID:
		return checker.NewDartagnan(mm), nil
	case checker.MockID:
		return checker.GetMock(), nil
	default:
		err := errors.New("error: unknown checker")
		return nil, verror(internalError, err)
	}
}
