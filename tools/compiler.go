// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

// Package tools contains a set of helper functions to compile programs with clang as well as
// wrappers to move, copy and remove files.
package tools

import (
	"os"
	"regexp"
	"strings"

	"vsync/logger"
)

func init() {
	RegEnv("CLANG_CMD", "clang",
		"Path to clang or space-separated command to run clang")
	RegEnv("LLVM_LINK_CMD", "llvm-link",
		"Path to llvm-link or space-separated command to run llvm-link")
	RegEnv("CFLAGS", "",
		"Flags passed to clang when compiling the target file")

	RegEnv("GENMC_CMD", "genmc", "Path to genmc binary")
	RegEnv("GENMC_INCLUDE_PATH", "",
		"Path to genmc headers, e.g., genmc.h")
}

func genmcIncludes() (incPath string) {
	defer func() {
		logger.Debugf("genmc includes directory is at %s\n", incPath)
		if _, err := os.Stat(incPath); os.IsNotExist(err) {
			logger.Fatalf("invalid genmc include path '%s'", incPath)
		}
	}()

	// check if the user set the path for genmc includes, this is useful when
	//  --model-checker-path is used with checker
	if incPath = GetEnv("GENMC_INCLUDE_PATH"); incPath != "" {
		logger.Debugf("GENMC_INCLUDE_PATH is set to=%s\n", incPath)
		return
	}

	// when GenMC runs with .ll file it prints the path the installed includes
	// so now we create an empty program, compile it to LLVM IR and call genmc
	const tinyProgram = `int main() { return 0; }`

	fn, err := Touch("vsyncer-tiny-*.c")
	if err != nil {
		logger.Fatalf("could not create temporary file: %v", err)
	}
	defer os.Remove(fn)

	err = os.WriteFile(fn, []byte(tinyProgram), 0644)
	if err != nil {
		logger.Fatalf("could not write to temporary file: %v", err)
	}

	clangCmd, err := FindCmd("CLANG_CMD")
	if err != nil {
		logger.Fatalf("could not find clang: %v", err)
	}

	var fnll = fn + ".ll"
	_, err = RunCmd(clangCmd[0], append(clangCmd[1:],
		"-S", "-emit-llvm", "-o", fnll, fn), nil)
	if err != nil {
		logger.Fatalf("could not run clang: %v", err)
	}
	defer os.Remove(fnll)

	genmcCmd, err := FindCmd("GENMC_CMD")
	if err != nil {
		logger.Fatalf("could not find genmc: %v", err)
	}

	output, err := RunCmd(genmcCmd[0], append(genmcCmd[1:], "--", fnll), nil)
	if err != nil {
		logger.Fatalf("could not run genmc: %v", err)
	}

	paths := regexp.MustCompile(`'-I (.*)'`).FindAllStringSubmatch(output, -1)
	if len(paths) != 2 {
		logger.Fatalf("unexpected number of paths in genmc message: %v", paths)
	}

	// There must be 2 paths reported by GenMC:
	// - the path where GenMC was built
	// - the path where GenMC is supposed to be installed
	//
	// We pick the first path that exists.
	//
	// The result of FindAllStringSubmatch is a list of pairs:
	//   [ [complete-match, ()-group], ...]
	//
	// We just want the second part of each pair.
	if FileExists(paths[0][1]) == nil {
		incPath = paths[0][1]
	} else {
		incPath = paths[1][1]
	}

	return
}

// Compile calls clang compiler and creates an LLVM IR module using the required compiler options.
func Compile(args []string, ofile string, addGenmcIncludePath bool) error {
	clang, err := FindCmd("CLANG_CMD")
	if err != nil {
		return err
	}

	var opts []string
	if cflags := GetEnv("CFLAGS"); cflags != "" {
		opts = append(opts, strings.Split(cflags, " ")...)
	}

	if addGenmcIncludePath {
		opts = append(opts,
			"-I", genmcIncludes(),
			"-D__CONFIG_GENMC_INODE_DATA_SIZE=64",
		)
	}

	opts = append(opts,
		"-DVSYNC_VERIFICATION",
		"-Xclang", "-disable-O0-optnone",
		"-g", "-S", "-emit-llvm",
		"-o", ofile,
	)

	opts = append(opts, args...)

	var (
		cmd     = clang[0]
		cmdArgs = append(clang[1:], opts...)
	)

	logger.Info("Compiling")
	logger.Infof("%v %v", cmd, strings.Join(cmdArgs, " "))
	out, err := RunCmd(cmd, cmdArgs, nil)
	logger.Debugf("%v", out)
	return err
}

type boilerplateMap map[string]string

func (bp boilerplateMap) names() []string {
	var names []string
	for k := range bp {
		names = append(names, k)
	}
	return names
}

func (bp boilerplateMap) stringNames() string {
	return strings.Join(bp.names(), " | ")
}
