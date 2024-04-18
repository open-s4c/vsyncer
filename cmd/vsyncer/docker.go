// Copyright (C) 2024 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"os"
	"vsync/tools"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const dockerDoc = `
Runs arguments in vsyncer Docker container.
`

var dockerFlags = struct {
	volumes []string
}{}

var dockerCmd = &cobra.Command{
	Use:   "vsyncer docker [flags] -- <command> [args]",
	Short: "Runs command in vsyncer Docker container",
	Long:  dockerDoc,
	RunE:  dockerRun,

	DisableFlagsInUseLine: true,
	SilenceUsage:          true,
	SilenceErrors:         true,
}

var dockerEmptyCmd = &cobra.Command{
	Use:   "docker [flags] -- <command> [args]",
	Short: "Runs command in vsyncer Docker container",
	RunE: func(cmd *cobra.Command, args []string) error {
		os.Args = os.Args[1:]
		return dockerCmd.Execute()
	},

	DisableFlagParsing:    true,
	DisableFlagsInUseLine: true,
}

func init() {
	dockerEmptyCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
			f.Hidden = true
		})
		cmd.Parent().HelpFunc()(cmd, args)
	})
	rootCmd.AddCommand(dockerEmptyCmd)
	flags := dockerCmd.Flags()
	flags.StringSliceVarP(&dockerFlags.volumes, "volume", "v", []string{}, "mount volumes")
}

func dockerRun(_ *cobra.Command, args []string) error {
	return tools.DockerRun(context.Background(), args, dockerFlags.volumes)
}
