// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

// Config enables multiple options when loading a LLVM IR module.
type Config struct {
	Atomics      bool     // whether we are selecting atomics/non-atomics
	EntryFunc    []string // a list of entry functions for analysis, default "run"
	Expand       bool     // whether it should expand the callgraph
	ExpandOnly   []string // a list of function prefixes to expand
	IdentifyOnly []string // a list of function prefixes to identify/atomify
	SkipFuncPref []string // a list of function prefixes to skip identify/atomify
	Args         []string // a list of arguments to pass to the command line of the checker
}

// DefaultConfig returns a default configuration for loading a module
func DefaultConfig() Config {
	return Config{
		EntryFunc:    []string{"main"},
		SkipFuncPref: []string{"pthread_", "__assert_fail", "llvm.", "_VERIFIER"},
		Expand:       true,
	}

}
