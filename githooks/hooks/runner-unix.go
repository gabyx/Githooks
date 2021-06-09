// +build !windows

package hooks

import cm "gabyx/githooks/common"

// GetDefaultRunner gets the default hook runner.
func GetDefaultRunner(hookPath string) cm.IExecutable {
	return &cm.Executable{
		Cmd:  "sh",
		Args: []string{hookPath}}
}
