// Copyright (C) 2024 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func dockerUserGroup(ctx context.Context) ([]string, error) {
	return nil, nil
}

func dockerInteractive(_ *exec.Cmd) error {
	return fmt.Errorf("functionality not supported")
}

// Convert Windows path to UNIX path
func ToSlash(path string) string {
	// fix path from Windows
	path = filepath.ToSlash(path)
	if len(path) <= 2 && !strings.Contains(path, ":") {
		return path
	}

	// This is most likely a string such as
	//     c:/something/something
	// Remove : and prepend a /
	path = strings.Replace(path, ":", "", 1)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}
