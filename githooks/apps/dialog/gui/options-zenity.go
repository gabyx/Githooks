//go:build !windows

package gui

import (
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

func getChoicesZenity(output string) (indices []uint) {
	out := strings.Split(strings.TrimSpace(output), "\x1e")

	indices = make([]uint, 0, len(out))

	for i := range out {
		idx, err := strconv.ParseUint(out[i], 10, 32) // nolint: mnd
		if err == nil {
			indices = append(indices, uint(idx))
		}
	}

	return
}

// ShowOptionsZenity shows a option dialog with `zenity`.
func ShowOptionsZenity(ctx context.Context, zenity string, opts *sets.Options) (r res.Options, err error) {
	if len(opts.Options) == 0 {
		err = cm.ErrorF("You need at least one option specified.")

		return r, err
	}

	if opts.Style == sets.OptionsStyleButtons && !opts.MultipleSelection {
		return showOptionsWithButtons(ctx, opts,
			func(ctx context.Context, m *sets.Message) (res.Message, error) {
				return ShowMessageZenity(ctx, zenity, m)
			})
	}

	args := []string{
		"--list",
		"--hide-header",
		"--column=id",
		"--column=",
		"--hide-column=1",
		"--print-column=1"}

	// Zenity prints default title and text if not set.
	args = append(args, "--title", opts.Title)
	args = append(args, "--text", opts.Text, "--no-markup")

	if opts.Width > 0 {
		args = append(args, "--width", strconv.FormatUint(uint64(opts.Width), 10))
	}

	if opts.Height > 0 {
		args = append(args, "--height", strconv.FormatUint(uint64(opts.Height), 10))
	}

	switch opts.WindowIcon {
	case sets.ErrorIcon:
		args = append(args, "--window-icon=error")
	case sets.WarningIcon:
		args = append(args, "--window-icon=warning")
	case sets.InfoIcon:
		args = append(args, "--window-icon=info")
	case sets.QuestionIcon:
		args = append(args, "--window-icon=question")
	}

	if strs.IsNotEmpty(opts.OkLabel) {
		args = append(args, "--ok-label", opts.OkLabel)
	}

	if strs.IsNotEmpty(opts.CancelLabel) {
		args = append(args, "--cancel-label", opts.CancelLabel)
	}

	if opts.ExtraButtons != nil {
		for i := range opts.ExtraButtons {
			if strs.IsEmpty(opts.ExtraButtons[i]) {
				return res.Options{}, cm.ErrorF("Empty label for extra button is not allowed")
			}

			args = append(args, "--extra-button", opts.ExtraButtons[i])
		}
	}

	if opts.NoWrap {
		args = append(args, "--no-wrap")
	}

	if opts.Ellipsize {
		args = append(args, "--ellipsize")
	}

	if opts.DefaultCancel {
		args = append(args, "--default-cancel")
	}

	// List options
	if opts.MultipleSelection {
		args = append(args, "--multiple")
		args = append(args, "--separator", "\x1e")
	}

	// Add choices with ids.
	for i := range opts.Options {
		args = append(args, strconv.Itoa(i), opts.Options[i])
	}

	out, err := gunix.RunZenity(ctx, zenity, args, "")
	if err == nil {
		return res.Options{
			General: res.OkResult(),
			Options: getChoicesZenity(string(out))}, nil
	}

	exErr := &exec.ExitError{}
	if errors.As(err, &exErr) {
		if exErr.ExitCode() == 1 {
			// Handle extra buttons.
			if len(out) > 0 {
				button := string(out[:len(out)-1])
				for i := range opts.ExtraButtons {
					if button == opts.ExtraButtons[i] {
						return res.Options{
							General: res.ExtraButtonResult(uint(i))}, nil
					}
				}
			}

			return res.Options{General: res.CancelResult()}, nil
		}
	}

	return res.Options{}, err
}
