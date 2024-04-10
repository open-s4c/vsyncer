// Copyright (C) 2024 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"vsync/logger"

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
	Run: func(cmd *cobra.Command, args []string) {
		os.Args = os.Args[1:]
		if err := dockerCmd.Execute(); err != nil {
			logger.Fatal(err)
		}
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

func startReader(r io.ReadCloser, w io.Writer) {
	scanner := bufio.NewScanner(r)
	go func() {
		for scanner.Scan() {
			fmt.Fprintln(w, scanner.Text())
		}
	}()
}

func startReaders(c *exec.Cmd) error {
	outReader, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	startReader(outReader, os.Stdout)
	errReader, err := c.StderrPipe()
	if err != nil {
		return err
	}
	startReader(errReader, os.Stderr)
	return nil
}

const (
	docker = "docker"
	image  = "ghcr.io/open-s4c/vsyncer"
	tag    = "latest"
)

func dockerRun(_ *cobra.Command, args []string) error {

	var (
		cmd = []string{"run", "--rm"}
	)
	// find current user
	u, err := user.Current()

	// find out current directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// check docker installation
	// is docker installed?
	//
	// are we running outside docker?
	//
	// is it rootless?
	// docker info -f "{{println .SecurityOptions}}" | grep rootless
	//
	// if not rooless do I have permission?
	// check if user in docker group, otherwise should we request sudo?

	// decide whether we should set uid:gid
	cmd = append(cmd, "-u", fmt.Sprintf("%v:%v", u.Uid, u.Gid))

	// mount current directory
	cmd = append(cmd, "-v", fmt.Sprintf("%s:%s", cwd, cwd))

	for _, v := range dockerFlags.volumes {
		cmd = append(cmd, "-v", fmt.Sprintf("%s:%s", v, v))
	}

	// set working directory to be current directory
	cmd = append(cmd, "-w", cwd)

	// docker image
	cmd = append(cmd, fmt.Sprintf("%s:%s", image, tag))

	// user arguments
	cmd = append(cmd, args...)

	// create command, start output readers and start
	c := exec.CommandContext(context.Background(), docker, cmd...)
	if err := startReaders(c); err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}
