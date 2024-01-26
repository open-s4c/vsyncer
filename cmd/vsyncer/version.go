// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/spf13/cobra"

	"vsync/logger"
)

var (
	name    = "vsyncer"
	version = "latest"
)

var versionCmd = cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Printf("%s %s\n", name, version)
	},
}

func register() {
	rootCmd.AddCommand(&versionCmd)
}

func init() {
	versionCmd.SetHelpFunc(func(command *cobra.Command, strings []string) {})
	register()
}
