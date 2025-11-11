//go:build !windows

package gui

import (
	"context"
	"errors"
	"os/exec"
	"path"
	"strconv"
	"strings"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// ShowFileSave shows a file-save dialog with `zenity`.
func ShowFileSaveZenity(ctx context.Context, zenity string, s *set.FileSave) (res.File, error) {
	args := []string{"--file-selection", "--save"}

	if strs.IsNotEmpty(s.Title) {
		args = append(args, "--title", s.Title)
	}

	if s.Width > 0 {
		args = append(args, "--width", strconv.FormatUint(uint64(s.Width), 10))
	}

	if s.Height > 0 {
		args = append(args, "--height", strconv.FormatUint(uint64(s.Height), 10))
	}

	switch s.WindowIcon {
	case set.ErrorIcon:
		args = append(args, "--window-icon=error")
	case set.WarningIcon:
		args = append(args, "--window-icon=warning")
	case set.InfoIcon:
		args = append(args, "--window-icon=info")
	case set.QuestionIcon:
		args = append(args, "--window-icon=question")
	}

	if strs.IsNotEmpty(s.Filename) {
		args = append(args, "--filename", s.Filename)
	}

	if s.OnlyDirectories {
		args = append(args, "--directory")
	}

	if s.ConfirmOverwrite {
		args = append(args, "--confirm-overwrite")
	}

	args = append(args, initFiltersZenity(s.FileFilters)...)

	out, err := gunix.RunZenity(ctx, zenity, args, s.Root)
	if err == nil {
		return res.File{
				General: res.OkResult(),
				Paths:   strings.Split(strings.TrimSpace(string(out)), "\x1e")},
			nil
	}

	exErr := &exec.ExitError{}
	if errors.As(err, &exErr) {
		if exErr.ExitCode() == 1 {
			return res.File{General: res.CancelResult()}, nil
		}
	}

	return res.File{}, err
}

// ShowFileSelection shows a file-selection dialog.
func ShowFileSelectionZenity(ctx context.Context, zenity string, s *set.FileSelection) (res.File, error) {
	args := []string{"--file-selection"}

	if strs.IsNotEmpty(s.Title) {
		args = append(args, "--title", s.Title)
	}

	if s.Width > 0 {
		args = append(args, "--width", strconv.FormatUint(uint64(s.Width), 10))
	}

	if s.Height > 0 {
		args = append(args, "--height", strconv.FormatUint(uint64(s.Height), 10))
	}

	switch s.WindowIcon {
	case set.ErrorIcon:
		args = append(args, "--window-icon=error")
	case set.WarningIcon:
		args = append(args, "--window-icon=warning")
	case set.InfoIcon:
		args = append(args, "--window-icon=info")
	case set.QuestionIcon:
		args = append(args, "--window-icon=question")
	}

	if strs.IsNotEmpty(s.Filename) || strs.IsNotEmpty(s.Root) {
		args = append(args, "--filename", path.Join(s.Root, s.Filename))
	}

	if s.OnlyDirectories {
		args = append(args, "--directory")
	}

	if s.MultipleSelection {
		args = append(args, "--multiple", "--separator", "\x1e")
	}

	args = append(args, initFiltersZenity(s.FileFilters)...)

	out, err := gunix.RunZenity(ctx, zenity, args, "")
	if err == nil {
		// Any linebreak at the end will be trimmed away.
		s := strings.TrimSuffix(string(out), "\n")

		return res.File{
				General: res.OkResult(),
				Paths:   strings.Split(s, "\x1e")},
			nil
	}

	exErr := &exec.ExitError{}
	if errors.As(err, &exErr) {
		if exErr.ExitCode() == 1 {
			return res.File{General: res.CancelResult()}, nil
		}
	}

	return res.File{}, err
}

func initFiltersZenity(filters []set.FileFilter) []string {
	var res []string
	for _, f := range filters {
		var buf strings.Builder
		buf.WriteString("--file-filter=")
		if f.Name != "" {
			buf.WriteString(f.Name)
			buf.WriteRune('|')
		}
		for _, p := range f.Patterns {
			buf.WriteString(p)
			buf.WriteRune(' ')
		}
		res = append(res, buf.String())
	}

	return res
}
