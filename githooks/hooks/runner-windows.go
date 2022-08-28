//go:build windows

package hooks

import (
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// ShellWrappedExecutable contains the data to a script/executable file which
// is wrapped over the shell with `-c "..."`.
type shellWrappedExecutable struct {
	Cmd string
	Env []string
}

// GetCommand gets the first command.
func (e *shellWrappedExecutable) GetCommand() string {
	return "sh"
}

// GetArgs gets all args.
func (e *shellWrappedExecutable) GetArgs(args ...string) (a []string) {
	cmd := e.Cmd

	for i := range args {
		cmd += strs.Fmt(" \"$%d\"", i+1)
	}

	// See https://pubs.opengroup.org/onlinepubs/009604499/utilities/sh.html
	return append([]string{"-c", cmd, "_"}, args...)
}

// GetArgs gets all args.
func (e *shellWrappedExecutable) GetString() string {
	return strs.Fmt("%s %q", e.Cmd, e.GetArgs())
}

// GetEnvironment returns all env. variables.
func (e *shellWrappedExecutable) GetEnvironment() []string {
	return envs
}

// GetDefaultRunner gets the default hook runner.
// Because you cannot set a hook as executable on Windows and
// there is no notion of a shebang in such an executable hook file
// we need to do some workaround.
// To make shebang work anyway we execute the hook inside the shell iteself
// (which wraps this shebang behavior on windows):
// Launch by `-c` and quote `hookPath`and all arguments into one argument only.
// This starts the shell and reads the shebang line on Windows.
// We assume here that a shell like git-bash.exe from https://gitforwindows.org/
// is installed where the `sh` is in the PATH when executing this hook over git.
func GetDefaultRunner(hookPath string, envs []string) cm.IExecutable {
	return &shellWrappedExecutable{Cmd: hookPath, Env: envs}
}
