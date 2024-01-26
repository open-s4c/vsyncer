// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"vsync/core"
	"vsync/logger"
	"vsync/module"
	"vsync/tools"
)

type bitseqFlag struct {
	short string
	long  string
	value string
	bs    core.Bitseq
}

var liftSelection = []core.Selection{
	core.SelectionLoads,
	core.SelectionStores,
}

var orderSelection = []core.Selection{
	core.SelectionAtomic,
	core.SelectionRMWs,
	core.SelectionFences,
}

var bitseqFlags = map[core.Selection]*bitseqFlag{
	core.SelectionLoads:  {"L", "loads", "", core.Bitseq{}},
	core.SelectionStores: {"S", "stores", "", core.Bitseq{}},
	core.SelectionAtomic: {"A", "atomics", "", core.Bitseq{}},
	core.SelectionRMWs:   {"X", "rmws", "", core.Bitseq{}},
	core.SelectionFences: {"F", "fences", "", core.Bitseq{}},
}

func addMutateFlags(flags *pflag.FlagSet) {
	for _, v := range bitseqFlags {
		flags.StringVarP(&v.value, v.long, v.short, "", fmt.Sprintf("bitseq for %s", v.long))
	}
	flags.SetInterspersed(false)
}

var mutateCmd = cobra.Command{
	Use:   "mutate [flags] <input.ll>",
	Short: "Mutate input file given a bitseq",
	Args:  IsArgsn,

	DisableFlagsInUseLine: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			outputGen = newOutputGenerator(args)
			fn        = outputGen("")
		)

		if hasToCompile(args) {
			if err := Compile(fn, args...); err != nil {
				return err
			}
			defer tools.Remove(fn)
		}

		m, err := mutate(fn, liftSelection, orderSelection)
		if err != nil {
			return err
		}
		defer m.Cleanup()

		ofn := rootFlags.outputFn
		logger.Debugf("Output file '%s'", ofn)
		tools.Dump(m, ofn)
		m.PrintSummary()
		return m.PrintDiff()
	},
}

func init() {
	rootCmd.AddCommand(&mutateCmd)
	addMutateFlags(mutateCmd.PersistentFlags())
}

func mutate(fn string, stages ...[]core.Selection) (*module.History, error) {
	var (
		cfg = moduleConfig()
		err error
	)

	m, err := module.Load(fn, cfg)
	if err != nil {
		return nil, verror(internalError, err)
	}

	for _, sgroup := range stages {
		for _, sel := range sgroup {
			bitseqSel := bitseqFlags[sel]
			if bitseqSel.value == "" {
				continue
			}
			a := m.Assignment(sel)
			a.Bs, err = core.ParseBitseq(bitseqSel.value, a.Bs.Length())
			if err != nil {
				return nil, verror(internalError, err)
			}
			logger.Debugf("Applying assignment %v", a)

			if err := m.Mutate(a); err != nil {
				return nil, verror(internalError, err)
			}

			if err := m.Record(); err != nil {
				return nil, verror(internalError, err)
			}
		}
	}
	return m, nil
}
