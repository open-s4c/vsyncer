// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package tools

import (
	"os"
	"strings"
)

// FindCmd looks for the value of an environment variable.
// If not set returns a default value.
func FindCmd(envVar string, defaultVal ...string) ([]string, error) {
	cmd, has := os.LookupEnv(envVar)
	if has {
		return strings.Split(cmd, " "), nil
	}
	return defaultVal, nil
}
