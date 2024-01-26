// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

// Package optimizer contains the optimization algorithms of vsyncer. Given a checker, optimization drivers
// search for a maximally-relaxed assignment for a module.
package optimizer

import (
	"context"
	"time"

	"vsync/checker"
	"vsync/core"
	"vsync/logger"
)

// Strategy defines the optimization strategy.
type Strategy int

const (
	// DDmin is a delta-debugging-based optimization algorithm
	DDmin Strategy = iota
	// LR is the linear relaxation algorithm of the VSync paper
	LR
)

// DriverConfig represents the configuration of the driver
type DriverConfig struct {
	BitsPerOp      int
	Filter         filterStrategy
	InitialBitseq  core.Bitseq
	GenmcOpts      []string
	Alpha          float64
	Tau            time.Duration
	Strategy       Strategy
	ErrorAsInvalid bool
}

// Driver is the object that coordinates the optimization
type Driver struct {
	cfg     DriverConfig
	atype   core.Selection
	checker checker.Tool
	filter  filterSet
	stats   *Stats
}

// NewDriver returns a new driver object
func NewDriver(cfg DriverConfig, c checker.Tool, stats *Stats) *Driver {
	return &Driver{
		cfg:     cfg,
		atype:   core.SelectionAtomic,
		checker: c,
		filter:  make(filterSet),
		stats:   stats,
	}
}

// Solution represents one possible correct assignment.
type Solution struct {
	bs      core.Bitseq
	status  checker.CheckStatus
	elapsed time.Duration
}

// Bitseq returns the bitsequence of the solution.
func (s Solution) Bitseq() core.Bitseq {
	return s.bs
}

// MutableModule is the module interface expected by any optimization driver.
type MutableModule interface {
	checker.DumpableModule
	Mutate(a core.Assignment) error
	Assignment(sel core.Selection) core.Assignment
}

func (d *Driver) recheck(ctx context.Context, m MutableModule, a core.Assignment, sol []Solution) (int, time.Duration) {
	// we are speculating, so double check most relaxed solution
	var (
		s      = sol[0]
		ts     = time.Now()
		status = s.status
		err    error
		r      checker.CheckResult
	)
	logger.Printf("RECHECK %v ", s.bs)

	if status == checker.CheckTimeout {
		if err = m.Mutate(core.Assignment{Bs: s.bs, Sel: a.Sel}); err != nil {
			logger.Fatal(err)
		}
		r, err = d.checker.Check(ctx, m)
		status = r.Status
	}

	elapsed := time.Since(ts)
	if status == checker.CheckOK {
		logger.Println("OK     ", elapsed)
		d.stats = NewStats()
		return 0, elapsed
	}

	logger.Println("FAIL   ", elapsed)
	d.stats.AddTime("failure", elapsed)

	// remember the bs failed
	d.filter.Set(s.bs)

	// select next initial bitseq: find most relaxed correct bs in
	// solutions slice or use initial bs.
	firstOK := pickNext(sol, checker.CheckOK)
	// since the initial bs is part of sol, firstOK is never -1
	if firstOK < len(sol) {
		// if if above is to make the linter happy
		a.Bs = sol[firstOK].bs
	}

	// index 0 failed the recheck, so we should only retry if there is
	// any timedout bitseq between [1; firstOK)
	if idx := pickNext(sol[1:firstOK], checker.CheckTimeout); idx == -1 {
		return firstOK, elapsed
	}

	return -1, elapsed
}

// Run starts the optimizer for a module with a given combination (bitsequence/selection).
func (d *Driver) Run(ctx context.Context, m MutableModule, at core.Selection) Solution {
	// if tau == 0, there is no speculation
	tau := d.cfg.Tau

	// initial assignment
	a := m.Assignment(at)

	logger.Println("== OPTIMIZATION ==============================")
	logger.Println()
	for {
		logger.Println("START  ", a.Bs, "#1 =", a.Bs.Ones(), time.Now().Format("15:04:05"))

		check := d.getCheckClosure(m, at, tau)
		var sol []Solution
		switch d.cfg.Strategy {
		case DDmin:
			sol = d.ddmin2(ctx, a.Bs, check, u2)
		case LR:
			sol = d.lr(ctx, a.Bs, check)
		default:
			logger.Fatal("unknown strategy")
		}

		// assume input is a correct solution
		sol = append(sol, Solution{bs: a.Bs, status: checker.CheckOK})
		logCurrentSolutions(sol)

		// if speculation is disabled, return most relaxed solution
		// if no new solution was found, return current most relaxed solution
		if s := sol[0]; s.bs.Equals(a.Bs) || tau == 0 {
			return s
		}

		idx, elapsed := d.recheck(ctx, m, a, sol)
		if idx > len(sol) {
			logger.Fatal("out of bound index in solutions")
		}
		if idx >= 0 && idx < len(sol) {
			return sol[idx]
		}

		// adapt tau to a higher value
		tau = adjustTau(tau, elapsed, d.cfg.Alpha)
		logger.Println("NEW TAU", tau)
	}
}

func adjustTau(tau time.Duration, elapsed time.Duration, alpha float64) time.Duration {
	if alpha == 0 {
		return tau + elapsed
	}
	if tau == 0 {
		return elapsed
	}
	c := 1000.0
	a := time.Duration(alpha * c)
	b := time.Duration(c) - a
	ntau := (a*tau + b*elapsed) / time.Duration(c)
	if ntau > tau {
		return ntau
	}
	return tau
}

func logCurrentSolutions(sol []Solution) {
	logger.Info("Current solutions")
	for _, s := range sol {
		logger.Info("+", s.bs, s.status)
	}
}

// pick most relaxed bs that we know is correct/timedout
func pickNext(sol []Solution, status checker.CheckStatus) int {
	var s Solution
	for i := 0; len(sol) > 0; i++ {
		s, sol = sol[0], sol[1:]
		if s.status == status {
			return i
		}
	}
	return -1
}

type checkClosure func(ctx context.Context, bs core.Bitseq) (checker.CheckStatus, time.Duration)

func (d *Driver) filterUpdate(bs core.Bitseq, status checker.CheckStatus, elapsed time.Duration) {
	switch status {
	case checker.CheckOK:
		logger.Println("OK     ", elapsed)
		d.stats.Inc(Success)
		d.stats.AddTime("success", elapsed)
	case checker.CheckTimeout:
		logger.Println("TIMEOUT", elapsed)
		d.stats.Inc(Timeout)
		d.stats.AddTime("timeout", elapsed)
	case checker.CheckNotSafe:
		logger.Println("NOTSAFE", elapsed)
		d.stats.Inc(NotSafe)
		d.stats.AddTime("failure", elapsed)
		d.filter.Set(bs)
	case checker.CheckNotLive:
		logger.Println("NOTLIVE", elapsed)
		d.stats.Inc(NotLive)
		d.stats.AddTime("failure", elapsed)
		d.filter.Set(bs)
	case checker.CheckInvalid:
		logger.Println("INVALID", elapsed)
		d.stats.Inc(Invalid)
		d.filter.Set(bs)
	default:
		logger.Fatal("unknown status")
	}
}

func (d *Driver) getCheckClosure(m MutableModule, at core.Selection, tau time.Duration) checkClosure {
	return func(ctx context.Context, bs core.Bitseq) (checker.CheckStatus, time.Duration) {
		logger.Printf("CHECK   %v ", bs)
		ts := time.Now()
		if tau > 0 {
			var cancel func()
			ctx, cancel = context.WithTimeout(ctx, tau)
			defer cancel()
		}

		if err := m.Mutate(core.Assignment{Bs: bs, Sel: at}); err != nil {
			elapsed := time.Since(ts)
			logger.Debugf("Failed mutation: %v", err)
			logger.Println("INVALID", elapsed)
			d.stats.Inc(Total)
			d.stats.Inc(Invalid)
			d.filter.Set(bs)
			return checker.CheckInvalid, elapsed
		}
		var (
			err error
			r   checker.CheckResult
		)
		r, err = d.checker.Check(ctx, m)
		status := r.Status

		elapsed := time.Since(ts)
		d.stats.Inc(Total)

		if err != nil {
			if !d.cfg.ErrorAsInvalid {
				logger.Println("ERROR")
				logger.Println("unexpected error: run debug output with -d")
				logger.Println("== ERROR =====================================")
				logger.Fatalf("%v\n", r.Output)
			}
			d.stats.Inc(Error)
			logger.Println("ERROR -> INVALID")
			d.filter.Set(bs)
		}

		d.filterUpdate(bs, status, elapsed)
		return status, elapsed
	}
}
