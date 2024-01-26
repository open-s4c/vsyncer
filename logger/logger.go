// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

// Package logger implements a simple logger with a few error levels.
package logger

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// Level represents the amount of detail in which the log is output.
type Level int

const (
	fatal Level = iota
	// ERROR only log errors
	ERROR
	// WARN only log warnings and errors
	WARN
	// INFO log information, warnings and errors
	INFO
	// DEBUG log as much as possible
	DEBUG
)

var (
	logger *bufio.Writer
	level  Level
)

func init() {
	f := bufio.NewWriter(os.Stdout)
	logger = f
}

// SetFileDescriptor sets the file descriptor to which the output is sent.
// If fd is nil, no output is shown.
func SetFileDescriptor(fd *os.File) {
	logger = bufio.NewWriter(fd)
}

// SetLevel reconfigures the error level of the logger.
func SetLevel(l Level) {
	level = l
}

// Fatal works as Error, but aborts the program.
func Fatal(args ...any) {
	Println(args...)
}

// Fatalf works as Errorf, but aborts the program.
func Fatalf(format string, args ...any) {
	Printf(format, args...)
	Println()
}

// Error works as fmt.Print, but it adds a newline at the end of the format string.
func Error(args ...any) {
	if logger == nil || level < INFO {
		return
	}
	Println(args...)
}

// Errorf works as fmt.Printf, but it adds a newline at the end of the format string.
func Errorf(format string, args ...any) {
	if logger == nil || level < INFO {
		return
	}
	Printf(format, args...)
	Println(logger)
}

// Warn works as fmt.Print when error level is WARN. It adds a newline at the end of the format string.
func Warn(args ...any) {
	if logger == nil || level < WARN {
		return
	}
	Println(args...)
}

// Warnf works as fmt.Printf when error level is WARN. It adds a newline at the end of the format string.
func Warnf(format string, args ...any) {
	if logger == nil || level < WARN {
		return
	}
	Printf(format, args...)
	Println()
}

// Info works as fmt.Print when error level is INFO. It adds a newline at the end of the format string.
func Info(args ...any) {
	if logger == nil || level < INFO {
		return
	}
	Println(args...)
}

// Infof works as fmt.Printf when error level is INFO. It adds a newline at the end of the format string.
func Infof(format string, args ...any) {
	if logger == nil || level < INFO {
		return
	}
	Printf(format, args...)
	Println()
}

// Debug works as fmt.Print when error level is DEBUG. It adds a newline at the end of the format string.
func Debug(args ...any) {
	if logger == nil || level < DEBUG {
		return
	}
	Println(args...)
}

// Debugf works as fmt.Printf when error level is DEBUG. It adds a newline at the end of the format string.
func Debugf(format string, args ...any) {
	if logger == nil || level < DEBUG {
		return
	}
	Printf(format, args...)
	Println()
}

// Print works as fmt.Print, but flushes the file descriptor.
func Print(args ...any) {
	fprint(args...)
}

// Println works as fmt.Println, but flushes the file descriptor.
func Println(args ...any) {
	fprintln(args...)
}

// Printf works as fmt.Printf, but flushes the file descriptor.
func Printf(format string, args ...any) {
	fprintf(format, args...)
}

var fstr = fmt.Sprintf

func fprint(args ...any) {
	if _, err := fmt.Fprint(logger, args...); err != nil {
		fail()
	}
	flush()
}
func fprintln(args ...any) {
	if _, err := fmt.Fprintln(logger, args...); err != nil {
		fail()
	}
	flush()
}
func fprintf(format string, args ...any) {
	fprint(fstr(format, args...))
}

func flush() {
	if logger.Flush() != nil {
		fail()
	}
}

func fail() {
	// use fatal instead of panic to make linter happy
	log.Fatal()
}
