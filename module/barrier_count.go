// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

type barrierCount struct {
	SeqCst  int
	Acquire int
	Release int
	Relaxed int
}
