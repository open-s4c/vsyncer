// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"vsync/checker"
	"vsync/logger"
	"vsync/tools"
)

func cr(s checker.CheckStatus) checker.CheckResult {
	return checker.CheckResult{
		Status: s,
	}
}

var testCases = []struct {
	r    checker.CheckResult
	err  error
	file bool
	verr bool
}{
	{r: cr(checker.CheckOK)},
	{r: cr(checker.CheckNotLive)},
	{r: cr(checker.CheckNotSafe)},
	{r: cr(checker.CheckInvalid)},
	{r: cr(checker.CheckTimeout)},
	{r: cr(checker.CheckRejected)},
	{
		err: errors.New("some error"),
	},
	{
		file: true,
		err:  verror(internalError, errors.New("some error")),
	},
}

func TestCheck(t *testing.T) {
	loadFiles()
	rootFlags.checker = "mock"
	cMock := checker.GetMock()

	// cleanup at end
	defer func() {
		cMock.Result = checker.CheckResult{}
		cMock.Err = nil
		tools.MockFileExistsErr = nil
	}()

	var fn string
	if len(testCFiles) > 0 {
		fn = testCFiles[0].path()
	}
	assert.NotEqual(t, fn, "")

	for i, tc := range testCases {
		i, tc := i, tc
		testName := fmt.Sprintf("%v", i)
		t.Run(testName, func(t *testing.T) {
			logger.Println("starting", testName)
			args := []string{fn}

			tools.MockFileExistsErr = nil
			if tc.file {
				tools.MockFileExistsErr = errors.New("file error")
			}
			cMock.Err = tc.err
			cMock.Result = tc.r

			// disable mutations
			liftSelection = nil
			orderSelection = nil

			err := checkRun(nil, args)
			if tc.r.Status == checker.CheckOK && tc.err == nil {
				assert.Nil(t, err)
				return
			} else if terr, tok := tc.err.(*vError); tok {
				assert.NotNil(t, err)
				er, ok := err.(*vError)
				assert.True(t, ok)
				assert.Equal(t, terr.typ, er.typ)
				assert.Equal(t, tc.r.Status, er.status)
				return
			}
			assert.NotNil(t, err)
		})
	}

}
