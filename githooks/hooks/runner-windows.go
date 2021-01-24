// +build windows

package hooks

import (
	cm "rycus86/githooks/common"
	strs "rycus86/githooks/strings"
	"strings"
)

// GetDefaultRunner gets the default hook runner.
// On Windows, executing with the default shell `sh` will only work for shell scripts
// since there is no notion of a shebang on windows. To make shebang work anyway
// we launch by `-c` and quote hookPath and all arguments into one argument only.
// This starts the shell and reads the shebang line on windows.
// We assume here that a shell like git-bash.exe from https://gitforwindows.org/
// is installed.
func GetDefaultRunner(hookPath string, args []string) cm.Executable {
	return cm.Executable{
		Cmd: "sh",
		Args: []string{"-c", hookPath + " " +
			strings.Join(
				strs.Map(args, func(s string) string { return "'%s'" }),
				" ")}}
}
