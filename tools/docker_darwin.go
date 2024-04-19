// Copyright (C) 2024 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package tools

import (
	"context"
	"fmt"
	"os/exec"
	"os/user"
)

func dockerUserGroup(ctx context.Context) ([]string, error) {
	// find current user
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("could not find current user: %v", err)
	}

	return []string{"-u", fmt.Sprintf("%v:%v", u.Uid, u.Gid)}, nil
}

func dockerInteractive(_ *exec.Cmd) error {
	return fmt.Errorf("functionality not supported")
}

func ToSlash(path string) string {
	return path
}
