// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

//go:build tools
// +build tools

package main

import (
	_ "github.com/ashanbrown/forbidigo"
	_ "golang.org/x/tools/cmd/stringer"
)
