//go:build !windwos && !darwin

package gui

import (
	"context"
	"os/exec"

	strs "github.com/gabyx/githooks/githooks/strings"
)

// GetZenityExecutable gets the installed `zenity` executable.
func GetZenityExecutable() string {
	for _, tool := range [3]string{"zenity", "qarma", "matedialog"} {
		path, err := exec.LookPath(tool)
		if err == nil {
			return path
		}
	}

	return "zenity"
}

// RunZenity runs the a Zenity executable.
func RunZenity(ctx context.Context, args []string, workingDir string) ([]byte, error) {

	zenity := GetZenityExecutable()
	var cmd *exec.Cmd

	if ctx != nil {
		cmd = exec.CommandContext(ctx, zenity, args...)

	} else {
		cmd = exec.Command(zenity, args...)
	}

	if strs.IsNotEmpty(workingDir) {
		cmd.Dir = workingDir
	}

	out, err := cmd.Output()

	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	return out, err
}
