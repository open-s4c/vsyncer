// Copyright (C) 2024 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

func dockerUserGroup(ctx context.Context) ([]string, error) {
	var rootless bool

	// find current user
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("could not find current user: %v", err)
	}

	// check docker installation
	if err := exec.CommandContext(ctx, dockerCmd).Run(); err != nil {
		return nil, fmt.Errorf("could not run docker: %v", err)
	}

	// is it rootless?
	if output, err := exec.CommandContext(ctx, dockerCmd, "info", "-f",
		"{{println .SecurityOptions}}").Output(); err != nil {
		return nil, fmt.Errorf("could not run docker: %v", err)
	} else {
		rootless = strings.Contains(string(output), "rootless")
	}

	// if not rooless do I have permission?
	if rootless || u.Uid != "0" {
		return nil, nil
	}

	// check if user in docker group, otherwise should we request sudo?
	if output, err := exec.CommandContext(ctx, "id", "-Gn").Output(); err != nil {
		return nil, fmt.Errorf("could get user groups: %v", err)
	} else if !strings.Contains(string(output), "docker") {
		return nil, fmt.Errorf("user is not in docker group")
	}

	return []string{"-u", fmt.Sprintf("%v:%v", u.Uid, u.Gid)}, nil
}

func dockerInteractive(c *exec.Cmd) error {
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
	}
	defer func() { _ = ptmx.Close() }()
	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH                        // Initial resize.
	defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	// NOTE: The goroutine will keep reading until the next keystroke before returning.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)
	return nil
}

func ToSlash(path string) string {
	return path
}

func FromSlash(path string) string {
	return path
}
