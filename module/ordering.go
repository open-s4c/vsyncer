// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package module

import (
	"github.com/llir/llvm/ir/enum"

	"vsync/core"
)

func fromAtomicOrdering(ao enum.AtomicOrdering) core.Ordering {
	switch ao {
	case enum.AtomicOrderingMonotonic:
		return core.Relaxed
	case enum.AtomicOrderingRelease:
		return core.Release
	case enum.AtomicOrderingAcquire:
		return core.Acquire
	case enum.AtomicOrderingSequentiallyConsistent:
		return core.SeqCst
	default:
		return core.Invalid
	}
}

func toAtomicOrdering(o core.Ordering) enum.AtomicOrdering {
	switch o {
	case core.Relaxed:
		return enum.AtomicOrderingMonotonic
	case core.Release:
		return enum.AtomicOrderingRelease
	case core.Acquire:
		return enum.AtomicOrderingAcquire
	case core.SeqCst:
		return enum.AtomicOrderingSequentiallyConsistent
	default:
		return enum.AtomicOrderingNone
	}
}
