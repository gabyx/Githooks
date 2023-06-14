package container

import (
	"fmt"

	strs "github.com/gabyx/githooks/githooks/strings"
)

// ContainerizedExecutable contains the data to a script/executable file.
type ContainerizedExecutable struct {
	containerType ContainerManagerType

	Cmd string // The command.

	ArgsPre  []string // The pre arguments.
	ArgsEnv  []string // Arguments which set environment values inside the containarized run.
	ArgsPost []string // The post arguments.
}

// GetCommand gets the first command.
func (e *ContainerizedExecutable) GetCommand() string {
	return e.Cmd
}

// GetArgs gets all args.
func (e *ContainerizedExecutable) GetArgs(args ...string) []string {
	var a []string
	a = append(a, e.ArgsPre...)
	a = append(a, e.ArgsEnv...)
	a = append(a, e.ArgsPost...)

	return a
}

// GetString gets all args.
func (e *ContainerizedExecutable) GetString() string {
	return strs.Fmt("%s %q", e.Cmd, e.GetArgs())
}

// GetEnvironment gets all environment variables.
func (e *ContainerizedExecutable) GetEnvironment() []string {
	return nil
}

// ApplyEnvironmentToArgs gets all environment variables.
// The input list is prefiltered for all Githooks exported variables.
func (e *ContainerizedExecutable) ApplyEnvironmentToArgs(env []string) {
	if e.containerType == ContainerManagerTypeV.Docker {
		fmt.Printf("Apply ARGS: %v", env)
		for i := range env {
			e.ArgsEnv = append(e.ArgsEnv, "-e", env[i])
		}
	} else {
		panic("Not implemented.")
	}
}
