package gui

import "context"

// BringToForground brings any process pid to foreground.
func BringToForground(ctx context.Context, pid int) error {
	type data struct {
		Pid int
	}

	d := data{Pid: pid}
	_, err := RunOSAScript(ctx, "bring-to-foreground", d, "")

	return err
}
