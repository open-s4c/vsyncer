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
	"path/filepath"
	"strings"
	"vsync/logger"
)

var (
	dockerCmd   = "docker"
	dockerImage = "ghcr.io/open-s4c/vsyncer"
	dockerTag   = "latest"
	useDocker   = "false"
)

func init() {
	RegEnv("VSYNCER_DOCKER", useDocker, "Use Docker container when calling clang, GenMC, Dat3M, etc")
	RegEnv("VSYNCER_DOCKER_IMAGE", dockerImage, "Docker image with clang, GenMC, Dat3M")
	RegEnv("VSYNCER_DOCKER_TAG", dockerTag, "Docker image tag")
	RegEnv("VSYNCER_DOCKER_VOLUMES", "", "Comma-separated list of additional volumes to mount")
}

func DockerPull(ctx context.Context) error {
	cmd := []string{"pull",
		fmt.Sprintf("%s:%s", GetEnv("VSYNCER_DOCKER_IMAGE"), GetEnv("VSYNCER_DOCKER_TAG")),
	}
	out, err := exec.CommandContext(ctx, dockerCmd, cmd...).CombinedOutput()
	fmt.Println(string(out))
	return err
}

func DockerRun(ctx context.Context, args []string, volumes []string) error {
	var (
		cmd = []string{"run", "--rm"}
	)

	// are we running outside docker?
	if FileExists("/.dockerenv") == nil {
		return fmt.Errorf("running inside docker. Set VSYNCER_DOCKER=false")
	}

	// get user/group flags
	if u, err := dockerUserGroup(ctx); err != nil {
		return err
	} else if len(u) > 0 {
		cmd = append(cmd, u...)
	}

	// find out current directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// mount current directory
	cmd = append(cmd, "-v", fmt.Sprintf("%s:%s", cwd, cwd))

	if v := GetEnv("VSYNCER_DOCKER_VOLUMES"); v != "" {
		volumes = append(volumes, strings.Split(v, ",")...)
	}

	for _, v := range volumes {
		if abs, err := filepath.Abs(v); err != nil {
			return fmt.Errorf("could not find volume path '%s': %v", v, err)
		} else {
			cmd = append(cmd, "-v", fmt.Sprintf("%s:%s", abs, abs))
		}
	}

	// better hostname
	cmd = append(cmd, "--hostname", "vsyncer")

	// set working directory to be current directory
	cmd = append(cmd, "-w", cwd)

	// docker opts
	if len(args) == 0 {
		cmd = append(cmd, "-it")
	}

	// docker image
	cmd = append(cmd, fmt.Sprintf("%s:%s", GetEnv("VSYNCER_DOCKER_IMAGE"), GetEnv("VSYNCER_DOCKER_TAG")))

	// user arguments
	cmd = append(cmd, args...)

	// log complete command
	logger.Debugf("%v\n", append([]string{dockerCmd}, cmd...))

	// create command, start output readers and start
	if len(args) != 0 {
		c := exec.CommandContext(ctx, dockerCmd, cmd...)
		if err := startReaders(c); err != nil {
			return err
		}
		if err := c.Start(); err != nil {
			return err
		}
		return c.Wait()
	}
	// if no commands, use pty
	// first set a better prompt
	cmd = append(cmd, "/bin/sh", "-c", "echo \"export PS1='\\h:\\w % '\" > /tmp/bashrc && env PS1='' bash --rcfile /tmp/bashrc")

	// create command and start pty (OS-dependent code)
	return dockerInteractive(exec.CommandContext(ctx, dockerCmd, cmd...))
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
	inWriter, err := c.StdinPipe()
	if err != nil {
		return err
	}
	startReader(os.Stdin, inWriter)

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
