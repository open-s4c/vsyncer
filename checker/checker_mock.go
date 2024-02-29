// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package checker

import (
	"context"
)

// Mock is a simple mock object for testing.
type Mock struct {
	Err    error
	Result CheckResult
}

var mock Mock

// GetMock return a Mock singleton.
func GetMock() *Mock {
	return &mock
}

// Check returns the desired check result and error
func (c *Mock) Check(_ context.Context, _ DumpableModule) (CheckResult, error) {
	var err error
	if c.Err != nil {
		err = c.Err
	}
	return c.Result, err
}

func (c *Mock) GetVersion() string {
	return "v0.0.0"
}
