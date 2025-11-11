//go:build !windwos

package gui

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"

	gmac "github.com/gabyx/githooks/githooks/apps/dialog/gui/darwin"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// GetZenityExecutable gets the installed `zenity` executable.
func GetZenityExecutable() (string, error) {
	for _, tool := range [3]string{"zenity", "qarma", "matedialog"} {
		path, err := exec.LookPath(tool)
		if err == nil {
			return path, nil
		}
	}

	return "", cm.Error("Zenity executable not found in system path.")
}

// RunZenity runs the a Zenity executable.
func RunZenity(ctx context.Context, zenity string, args []string, workingDir string) (b []byte, err error) {
	var cmd *exec.Cmd
	handleErr := func() {
		if ctx != nil && ctx.Err() != nil {
			err = ctx.Err()
		}
	}

	if ctx != nil {
		cmd = exec.CommandContext(ctx, zenity, args...)
	} else {
		cmd = exec.Command(zenity, args...)
	}

	if strs.IsNotEmpty(workingDir) {
		cmd.Dir = workingDir
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err = cmd.Start()
	handleErr()
	if err != nil {
		return nil, err
	}

	if runtime.GOOS == "darwin" {
		// On macOS the zenity window needs to be brought to
		// the foreground.
		err = gmac.BringToForground(ctx, cmd.Process.Pid)
		handleErr()

		if err != nil {
			return nil, err
		}
	}

	err = cmd.Wait()
	handleErr()

	return stdout.Bytes(), err
}
