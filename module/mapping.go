// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"vsync/core"
	"vsync/logger"
)

func mapInstruction(in wrapInstruction) core.AtomicOp {
	switch in.(type) {
	case *wrapInstFence:
		return core.Fence
	case *wrapInstLoad:
		return core.Load
	case *wrapInstStore:
		return core.Store
	case *wrapInstAtomicRMW:
		return core.RMW
	case *wrapInstCmpXchg:
		return core.Cmpxchg
	default:
		logger.Fatalf("unknown type: %T", in)
	}
	return core.InvalidOp
}

func mapOrdering(in wrapInstruction, val int) core.Ordering {
	return mapInstruction(in).GetOrdering(val)
}
