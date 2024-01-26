// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"vsync/checker"
	"vsync/core"
	"vsync/logger"
	"vsync/module"
	"vsync/optimizer"
	"vsync/tools"
)

const cpuFactor = 2

var optimizeCmd = cobra.Command{
	Use:   "optimize [flags] <input.ll|input.c>",
	Short: "Finds an optimization for input file",
	Args:  IsArgsn,
	RunE:  optimizeRun,

	DisableFlagsInUseLine: true,
}

var optimizeFlags = struct {
	algorithm    string
	adaptive     bool
	timeout      time.Duration
	filter       string
	alpha        float64
	errorInvalid bool
}{}

func initOptimize() {
	rootCmd.AddCommand(&optimizeCmd)
	flags := optimizeCmd.PersistentFlags()
	addMutateFlags(flags)
	addCheckFlags(flags)
	flags.StringVarP(&optimizeFlags.algorithm, "algorithm", "a", "lr", "optimization algorithm (lr|ddmin)")
	flags.BoolVar(&optimizeFlags.errorInvalid, "error-as-invalid", false, "map checker errors as invalid mutations")
	flags.BoolVar(&optimizeFlags.adaptive, "adaptive", true, "use adaptive timeout to optimize")
	flags.DurationVar(&optimizeFlags.timeout, "speculate", 0, "speculate variant correct after given timeout")
	flags.StringVar(&optimizeFlags.filter, "filter", "rlx", "filter (none/dup/rlx)")
	flags.Float64Var(&optimizeFlags.alpha, "alpha", 0, "memory alpha for adaptive")
}

func compileConditional(fn string, args []string) (bool, error) {
	if !hasToCompile(args) {
		return false, nil
	}
	return true, Compile(fn, args...)
}

func optimizeRun(_ *cobra.Command, args []string) error {

	var (
		outputGen = newOutputGenerator(args)
		fn        = outputGen("")
		mm        = checker.ParseMemoryModel(checkFlags.memoryModel)
	)

	if remove, err := compileConditional(fn, args); err != nil {
		return err
	} else if remove {
		defer tools.Remove(fn)
	}

	m, err := mutate(fn, liftSelection, orderSelection)
	if err != nil {
		return err
	}
	defer m.Cleanup()

	if err := m.Record(); err != nil {
		return verror(internalError, err)
	}

	chkr, err := newChecker(parseCheckerID(rootFlags.checker), mm)
	if err != nil {
		return err
	}

	cfg := newDriverConfig()
	sts := optimizer.NewStats()
	sel := core.SelectionAtomic
	ia := m.Assignment(sel)
	d := optimizer.NewDriver(cfg, chkr, sts)
	s := d.Run(context.Background(), m, sel)
	defer logger.Println(sts)

	return evaluateOptimizeResult(s, chkr, m, ia)
}

func evaluateOptimizeResult(s optimizer.Solution, chkr checker.Tool, m *module.History, ia core.Assignment) error {
	// if the solution is the same as the input, we should check if the
	// user hasn't given a rather incorrect bs:
	if s.Bitseq().Equals(ia.Bs) {
		logger.Printf("RECHECK %v ", ia.Bs)
		ts := time.Now()
		if err := m.Mutate(ia); err != nil {
			return verror(internalError, err)
		}

		r, err := chkr.Check(context.Background(), m)
		elapsed := time.Since(ts)
		if err != nil {
			logger.Fatal(err)
		}

		switch r.Status {
		case checker.CheckOK:
			logger.Println("OK     ", elapsed)
			printSolutions(m, ia.Bs, s, true)

		default:
			logger.Println("FAIL   ", elapsed)
			printSolutions(m, ia.Bs, s, false)
		}
	} else {
		printSolutions(m, ia.Bs, s, true)
	}

	return nil
}

func newDriverConfig() optimizer.DriverConfig {
	cfg := optimizer.DriverConfig{
		ErrorAsInvalid: optimizeFlags.errorInvalid,
		Alpha:          optimizeFlags.alpha,
		BitsPerOp:      2,
	}
	if optimizeFlags.adaptive {
		cfg.Tau = 1 * time.Millisecond
	}
	if optimizeFlags.timeout != 0 {
		cfg.Tau = optimizeFlags.timeout

	}
	switch optimizeFlags.filter {
	case "none":
		cfg.Filter = optimizer.None
	case "dup":
		cfg.Filter = optimizer.Dup
	case "rlx":
		cfg.Filter = optimizer.Rlx
	default:
		logger.Fatalf("unknown filter type %v", optimizeFlags.filter)
	}

	switch optimizeFlags.algorithm {
	case "lr":
		cfg.Strategy = optimizer.LR
	case "ddmin":
		cfg.Strategy = optimizer.DDmin
	default:
		logger.Fatal("invalid algorithm", optimizeFlags.algorithm)
	}

	return cfg
}

func printSolutions(m *module.History, initial core.Bitseq, s optimizer.Solution, correct bool) {

	logger.Println()
	m.PrintSummary()
	if s.Bitseq().Equals(initial) && correct {
		logger.Printf("Result\n   No optimization found!\n")
		logger.Println()
	} else if s.Bitseq().Equals(initial) && !correct {
		logger.Printf("Result\n   Input bitseq incorrect!\n")
		logger.Println()
	} else {
		if err := m.Forget(); err != nil {
			logger.Fatal(err)
		}
		if err := m.Mutate(core.Assignment{Bs: s.Bitseq(), Sel: core.SelectionAtomic}); err != nil {
			logger.Fatal(err)
		}

		logger.Printf("Result\n   Optimization found!\n")
		logger.Println()
		if err := m.PrintDiff(); err != nil {
			logger.Println(err)
		}
	}
	logger.Println("== ITERATION STATS ===========================")
}

func defaultInstances(nb uint) uint {
	if nb != 0 {
		return nb
	}
	cpus := uint(runtime.NumCPU())
	if cpus == 1 {
		return 1
	}
	// use half of CPUs for random scheduling
	return cpus / cpuFactor
}
