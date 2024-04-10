// Copyright (C) 2023 Huawei Technologies Co., Ltd. All rights reserved.
// SPDX-License-Identifier: MIT

package tools

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"vsync/logger"
)

// FindCmd looks for the value of an environment variable.
// If not set returns a default value.
func FindCmd(key string) ([]string, error) {
	val, err := LookupEnv(key)
	if err != nil {
		return nil, err
	}

	cmds := strings.Split(val, " ")
	if IsDefaultEnv(key) && GetEnv("VSYNCER_DOCKER") == "true" {
		return append([]string{"vsyncer", "docker", "--"}, cmds...), nil
	}
	return cmds, err
}

type Envvar struct {
	Name string
	Desc string
	Defv string
}

func (e Envvar) String() string {
	return fmt.Sprintf("%s: %s (default: %s)", e.Name, e.Desc, e.Defv)
}

var envVars = map[string]Envvar{}

// RegEnv registers an environment variable with a default value and a
// description.
func RegEnv(key, defv, desc string) {
	if _, has := envVars[key]; has {
		logger.Fatalf("Envvar '%s' already registered", key)
	}
	envVars[key] = Envvar{
		Name: key,
		Defv: defv,
		Desc: desc,
	}
}

// GetEnv returns the value of an environment variable if available,
// otherwise returns its default value. If the environment variable
// was not registered, a fatal error is raised.
func GetEnv(key string) string {
	if val, err := LookupEnv(key); err != nil {
		log.Fatalf("%v", err)
		return ""
	} else {
		return val
	}
}

// LookupEnv returns the value of an environment variable if available,
// otherwise returns its default value. If the environment variable
// was not registered, an error is returned.
func LookupEnv(key string) (string, error) {
	if v, has := envVars[key]; !has {
		return "", fmt.Errorf("Envvar '%s' not registered", key)
	} else if vv, has := os.LookupEnv(key); has { //permit:os.LookupEnv
		return vv, nil
	} else {
		return v.Defv, nil
	}
}

// IsDefaultEnv returns true if the variable value is default, ie, the variable
// was not set by the user.
func IsDefaultEnv(key string) bool {
	if _, has := envVars[key]; !has {
		logger.Fatalf("Envvar '%s' not registered", key)
	}

	_, has := os.LookupEnv(key) //permit:os.LookupEnv
	return !has
}

// GetEnvvars returns list of all registered environment variables.
func GetEnvvars() []Envvar {
	var (
		keys []string
		envv []Envvar
	)
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		envv = append(envv, envVars[k])
	}
	return envv
}
