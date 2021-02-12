// +build linux

package gui

import (
	"context"
	"os/exec"
)

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
func RunZenity(ctx context.Context, args []string) ([]byte, error) {

	zenity := GetZenityExecutable()

	if ctx != nil {
		out, err := exec.CommandContext(ctx, zenity, args...).Output()
		if ctx.Err() != nil {
			err = ctx.Err()
		}

		return out, err
	}

	return exec.Command(zenity, args...).Output()
}
