//go:build !windows

package hooks

import cm "github.com/gabyx/githooks/githooks/common"

// GetDefaultRunner gets the default hook runner.
func GetDefaultRunner(hookPath string, envs []string) cm.IExecutable {
	return &cm.Executable{
		Cmd:  "sh",
		Args: []string{hookPath},
		Env:  cm.CopySlice(envs),
	}
}
