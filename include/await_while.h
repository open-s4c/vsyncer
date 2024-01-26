// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

#ifndef AWAIT_WHILE_H
#define AWAIT_WHILE_H

#include <stdbool.h>
void __VERIFIER_loop_begin(void);
void __VERIFIER_spin_start(void);
void __VERIFIER_spin_end(bool);

#define await_while(cond)                                                     \
    for (__VERIFIER_loop_begin();                                             \
         (__VERIFIER_spin_start(), (cond) ? 1 : (__VERIFIER_spin_end(1), 0)); \
         __VERIFIER_spin_end(0))
#endif
