// Copyright (C) 2024 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package tools

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

var (
	dockerPath  = "docker"
	dockerImage = "ghcr.io/open-s4c/vsyncer"
	dockerTag   = "latest"
	useDocker   = "false"
)

func init() {
	RegEnv("VSYNCER_DOCKER", useDocker, "Use vsyncer Docker container for GenMC, Dat3m, etc")
}

func DockerRun(ctx context.Context, args []string, volumes []string) error {
	var (
		cmd      = []string{"run", "--rm"}
		rootless = false
	)
	// find current user
	u, err := user.Current()

	// find out current directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// check docker installation
	if err := exec.CommandContext(ctx, dockerPath).Run(); err != nil {
		return fmt.Errorf("could not run docker: %v", err)
	}

	// are we running outside docker?
	if FileExists("/.dockerenv") == nil {
		return fmt.Errorf("running inside docker. Set VSYNCER_DOCKER=false")
	}

	// is it rootless?
	if output, err := exec.CommandContext(ctx, dockerPath, "info", "-f",
		"{{println .SecurityOptions}}").Output(); err != nil {
		return fmt.Errorf("could not run docker: %v", err)
	} else {
		rootless = strings.Contains(string(output), "rootless")
	}

	// if not rooless do I have permission?
	if !rootless && u.Uid != "0" {
		// check if user in docker group, otherwise should we request sudo?
		if output, err := exec.CommandContext(ctx, "id", "-Gn").Output(); err != nil {
			return fmt.Errorf("could get user groups: %v", err)
		} else if !strings.Contains(string(output), "docker") {
			return fmt.Errorf("user is not in docker group")
		}

		cmd = append(cmd, "-u", fmt.Sprintf("%v:%v", u.Uid, u.Gid))
	}

	// mount current directory
	cmd = append(cmd, "-v", fmt.Sprintf("%s:%s", cwd, cwd))

	for _, v := range volumes {
		cmd = append(cmd, "-v", fmt.Sprintf("%s:%s", v, v))
	}

	// set working directory to be current directory
	cmd = append(cmd, "-w", cwd)

	// docker image
	cmd = append(cmd, fmt.Sprintf("%s:%s", dockerImage, dockerTag))

	// user arguments
	cmd = append(cmd, args...)

	// create command, start output readers and start
	c := exec.CommandContext(ctx, dockerPath, cmd...)
	if err := startReaders(c); err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
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
