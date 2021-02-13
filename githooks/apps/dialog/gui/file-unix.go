// +build !windows,!darwin

package gui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	gunix "gabyx/githooks/apps/dialog/gui/unix"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	strs "gabyx/githooks/strings"
)

func ShowFileSave(ctx context.Context, s *set.FileSave) (res.File, error) {

	args := []string{"--file-selection", "--save"}

	if strs.IsNotEmpty(s.Title) {
		args = append(args, "--title", s.Title)
	}

	if s.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", s.Width))
	}

	if s.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", s.Height))
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

	args = append(args, initFilters(s.FileFilters)...)

	out, err := gunix.RunZenity(ctx, args, s.Root)
	if err == nil {
		return res.File{
				General: res.OkResult(),
				Paths:   strings.Split(strings.TrimSpace(string(out)), "\x00")},
			nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {
			return res.File{General: res.CancelResult()}, nil
		}
	}

	return res.File{}, err
}

func ShowFileSelection(ctx context.Context, s *set.FileSelection) (res.File, error) {

	args := []string{"--file-selection"}

	if strs.IsNotEmpty(s.Title) {
		args = append(args, "--title", s.Title)
	}

	if s.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", s.Width))
	}

	if s.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", s.Height))
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

	if s.MultipleSelection {
		args = append(args, "--multiple", "--separator", "\x1e")
	}

	args = append(args, initFilters(s.FileFilters)...)

	out, err := gunix.RunZenity(ctx, args, s.Root)
	if err == nil {
		return res.File{
				General: res.OkResult(),
				Paths:   strings.Split(strings.TrimSpace(string(out)), "\x1e")},
			nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {
			return res.File{General: res.CancelResult()}, nil
		}
	}

	return res.File{}, err
}

func initFilters(filters []set.FileFilter) []string {
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
