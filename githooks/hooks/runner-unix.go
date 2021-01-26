// +build !windows

package hooks

import cm "gabyx/githooks/common"

// GetDefaultRunner gets the default hook runner.
func GetDefaultRunner(hookPath string, args []string) cm.Executable {
	return cm.Executable{
		Cmd:  "sh",
		Args: append([]string{hookPath}, args...)}
}
