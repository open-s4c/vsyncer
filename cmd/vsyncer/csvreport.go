// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"vsync/checker"
	"vsync/logger"
)

const fileMode = 0600

type csvReport struct {
	name          string
	checker       checkerID
	memoryModel   checker.MemoryModel
	duration      time.Duration
	status        checker.CheckStatus
	version       string
	numExecutions int
	err           error
}

const (
	dateTime = "2006-01-02 15:04:05"
)

func (csv csvReport) save(filename string) {
	if filename == "" {
		return
	}
	withHeader := false
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		withHeader = true
	}

	fp, err := os.OpenFile(filename,
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, fileMode)
	if err != nil {
		logger.Fatalf("could not open file: %v", filename)
	}
	defer func() {
		if err := fp.Close(); err != nil {
			logger.Warnf("error closing file: %v", err)
		}
	}()

	if withHeader {
		fmt.Fprint(fp, "# date, filename, checker, version, memory_model, duration, status, num_executions, error_type, exit_code")
		fmt.Fprintln(fp)
	}

	fmt.Fprintf(fp, "%s, %s, %v, %v, %v, %v, %v, %d, %s, %d\n",
		time.Now().Format(dateTime),
		csv.name,
		csv.checker,
		csv.version,
		csv.memoryModel,
		csv.duration,
		csv.status,
		csv.numExecutions,
		getErrorType(csv.err),
		getErrorCode(csv.err))
}
