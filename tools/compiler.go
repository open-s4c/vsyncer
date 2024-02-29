// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

// Package tools contains a set of helper functions to compile programs with clang as well as
// wrappers to move, copy and remove files.
package tools

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"vsync/logger"
)

// Compile calls clang compiler and creates an LLVM IR module using the required compiler options.
func Compile(args []string, ofile string, addGenmcIncludePath bool) error {
	clang, err := FindCmd("CLANG_CMD", "clang")
	if err != nil {
		return err
	}

	var opts []string
	if cflags, has := os.LookupEnv("CFLAGS"); has {
		opts = append(opts, strings.Split(cflags, " ")...)
	}

	genmcIncludes := ""
	if addGenmcIncludePath {
		/* check if the user set the path for genmc includes, this is useful when --model-checker-path is used with checker */
		if envIncPath, has := os.LookupEnv("VSYNCER_GENMC_INCLUDE_PATH"); has {
			genmcIncludes = envIncPath
			logger.Debugf("VSYNCER_GENMC_INCLUDE_PATH is set to=%s\n", genmcIncludes)
		} else {
			genmc_path, err := exec.LookPath("genmc")
			if err != nil {
				log.Fatal("genmc was not found in PATH")
			}
			logger.Debugf("genmc is available at %s\n", genmc_path)
			install_base := filepath.Dir(filepath.Dir(genmc_path))
			genmcIncludes = filepath.Join(install_base, "include", "genmc")

			if _, err := os.Stat(genmcIncludes); os.IsNotExist(err) {
				log.Fatal("Unable to find genmc include directory.")
			}
		}

		logger.Debugf("genmc includes directory is at %s\n", genmcIncludes)
		opts = append(opts,
			"-I", genmcIncludes,
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
