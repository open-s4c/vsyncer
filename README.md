# vsyncer --- verifying and optimizing concurrent code on WMM

`vsyncer` is a toolkit to verify and optimize concurrent C/C++ programs on Weak
Memory Models (WMM).  The correctness of the target program is verified with
state-of-the-art model checkers [Dartagnan][] and [GenMC][].  Optimization
of the memory ordering of atomic operations is achieved by speculative
verification of LLVM-IR mutations of the target program with feedback from
the model checkers.  Our ASPLOS'21 [publication][] describes the motivation
and research effort put in this tool.  See references below for further
works using `vsyncer`.

The accompaigning [libvsync][] library contains several practical and efficient
data structures and synchronization primitives verified and optimized with
`vsyncer`.

## Installation

### Runtime dependencies

To run vsyncer you'll need the following tools:

- clang and llvm >= v10
- Dartagnan >= v4.0.0 (alternative)
- GenMC >= v0.8 (alternative)

You can have them installed on your system or use docker containers.

### Build from source

Beyond the runtime dependencies, you'll need Golang >= v1.18 to build
`vsyncer` from source.

Assuming `go` and `make` are on your path, simply run

    make
    cp build/vsyncer /directory/on/your/path

If you do not have make installed, simply build it with

    go build -o vsyncer ./cmd/vsyncer
    cp vsyncer /directory/on/your/path

## Overview

The `vsyncer` program offers several commands to manipulate and inspect
concurrent programs.  Central to `vsyncer` is the concept of **atomic
operations**, which encompass atomic memory operations such as atomic
reads, atomic writes, and read-modify-write operations as well as memory
fences. Every atomic operation has a **memory ordering** (called [barrier
mode][publication]). The memory ordering speficies how an atomic operation is
ordered in relation to other concurrent operations in the program.  `vsyncer`
supports four memory orderings: `SeqCst`, `Release`, `Acquire`, and `Relaxed`.

The input of `vsyncer` is an LLVM-IR module of compiled C/C++ userspace
program. If the program is not yet compiled, `vsyncer` calls `clang`
underneath to generate the `.ll` file.

Given a module, `vsyncer` statically analyzes its callgraph starting from the
`main()` function and keeps track of all operations that may be executed
in runtime.  A **selection** is the sequence of tracked operations in one
of the following kinds:

- *L*: all read operations (atomic and plain)
- *S*: all write operations (atomic and plain)
- *A*: all atomic operations
- *X*: the read-modify-write subset of atomic operations
- *F*: the memory fence subset of atomic operations

`vsyncer` is able to mainly perform two kinds of **mutations** (ie, program
transformations): (1) with *A*, *X*, and/or *F* selections, `vsyncer`
can modify the memory ordering of atomic operations, making them weaker
or stronger; or (2) with *L* or *S* selections, it can transform plain
operations (ie, ordinary non-atomic reads and writes) into atomic operations
and vice-versa.

An **assignment** is a selection of operations and a sequence
of values representing the operations' memory orderings or their plain/atomic
modifier (depending on the selection type).  The sequence of values is encoded
as a **bitsequence**  such as `0b001101` or `0x1a40`.

*L* and *S* assignments take bitsequences in which each bit represents whether a specific read or write operation is an atomic or plain operation.
*A*, *X*, and *F* assignments take bitsequences in which each **pair** of bits represent the memory ordering of a specific atomic operation.
The memory may be relaxed (`0b00`),  release (`0b01`), acquire (`0b10`), or sequentially consistent (`0b11`).

### Quick start

#### Installation

    go build -o /dir/in/path/vsyncer ./cmd/vsyncer

#### Retrieving information from program

    vsyncer info example/ttaslock.c

#### Checking whether program is correct

    vsyncer check example/ttaslock.c

#### Mutating program with a memory ordering assignment:

    vsyncer mutate -o ttaslock.ll example/ttaslock.c -A 0x123
    vsyncer info ttaslock.ll

To make all memory orderings be sequential consistent, you have to set all
the bits of the bitsequences. Since this is used quite often, `vsyncer`
accepts -1 as a shortcut for a bitsequence with all bits set.

#### Checking mutation

    vsyncer check ttaslock.ll

Or simply mutate and check in a single command:

    vsyncer check -A 0x123 example/ttaslock.c

Finally, to optimize the barriers, use `vsyncer optimze`. We suggest mutating
the program to have all atomic operations with sequential consistent memory
ordering first.

    vsyncer optimize -A -1 example/ttaslock.c


## Limitations

### Function pointers

Function pointers are not analyzed. The model checker still should guarantee
the correctness of the given program with function pointers, however, the
optimization does not consider them.  If a function (passed as pointer)
uses **too weak** memory orderings, `vsyncer optimize` won't be able to
fix the missing barriers, but the model checker should still report errors
if correctness is affected.  If a function (passed as pointer) user **too
strong** memory orderings, `vsyncer optimize` won't be able to optimize the
memory orderings, but the code will still be correct.


## Publications using `vsyncer`

- [VSync: push-button verification and optimization for synchronization primitives on weak memory models](https://dl.acm.org/doi/10.1145/3445814.3446748) --- ASPLOS'21, Oberhauser et al.
- [Verifying and Optimizing the HMCS Lock for Arm Servers]() --- NETYS'21, Oberhauser et al.
- [Verifying and Optimizing Compact NUMA-Aware Locks on Weak Memory Models](https://arxiv.org/abs/2111.1524) --- Technical report, 2022, Paolillo et al.
- [CLoF: A Compositional Lock Framework for Multi-level NUMA Systems](https://dl.acm.org/doi/10.1145/3477132.3483557) --- SOSP'22, Chehab et al.
- [BBQ: A Block-based Bounded Queue for Exchanging Data and Profiling](https://www.usenix.org/conference/atc22/presentation/wang-jiawei) --- ATC'22, Wang et al
- [BWoS: Formally Verified Block-based Work Stealing for Parallel Processing](https://www.usenix.org/conference/osdi23/presentation/wang-jiawei) --- OSDI'23, Wang et al.
- [AtoMig: Automatically Migrating Millions Lines of Code from TSO to WMM](https://dl.acm.org/doi/abs/10.1145/3575693.3579849) --- ASPLOS'23, Beck et al.

## License

`vsyncer` is released under the [MIT](LICENSE) license.

The dependencies have the following licenses:

| Package | Licence |
| --- | --- |
| github.com/fatih/color | MIT |
| github.com/jinzhu/copier | MIT |
| github.com/llir/ll | 0BSD |
| github.com/llir/llvm | 0BSD |
| github.com/llir/llvm/internal/natsort | MIT |
| github.com/mattn/go-colorable | MIT |
| github.com/mattn/go-isatty | MIT |
| github.com/mewmew/float | Unlicense |
| github.com/pkg/errors | BSD-2-Clause |
| github.com/spf13/cobra | Apache-2.0 |
| github.com/spf13/pflag | BSD-3-Clause |
| golang.org/x/sync/errgroup | BSD-3-Clause |
| golang.org/x/sys/unix | BSD-3-Clause |

Use [go-licenses](https://github.com/google/go-licenses) to review the licenses.


## Development information and contact

Directory structure is as follows:

- `core`:  most basic concepts such as atomic operation identifiers, memory ordering identifiers, selections, bit sequences, and assignments.
- `module`: an LLVM IR module, with all operations: load, mutate, diff, dump
- `checker`: takes module and checks with model checker
- `optimizer:` takes checker, takes module, optimizes
- `cmd/vsyncer`: the main function with subcommands

For questions write to `vsync AT huawei DOT com`.

This project is under the support of [OpenHarmony Concurrency & Coordination TSG (Technical Support Group), 并发与协同TSG][tsg].

[tsg]: https://www.openharmony.cn/techCommittee/aboutTSG
[publication]: https://dl.acm.org/doi/abs/10.1145/3445814.3446748
[Dartagnan]: https://github.com/hernanponcedeleon/Dat3M
[GenMC]: https://github.com/MPI-SWS/genmc
[libvsync]:  https://gitee.com/s4c/libvsync
