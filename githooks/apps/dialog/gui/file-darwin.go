//go:build darwin

package gui

import (
	"context"
	"os/exec"
	"strings"

	gmac "github.com/gabyx/githooks/githooks/apps/dialog/gui/darwin"

	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

func initFiltersOSAScript(filters []sets.FileFilter) []string {
	var filter []string
	for _, f := range filters {
		for _, p := range f.Patterns {
			star := strings.LastIndexByte(p, '*')
			if star >= 0 {
				dot := strings.LastIndexByte(p, '.')
				if star > dot {
					return nil // we got ".*" -> return no filter
				}

				filter = append(filter, p[dot+1:]) // append the *.()
			} else {
				filter = append(filter, p)
			}
		}
	}

	return filter
}

func translateFileSelection(f *sets.FileSelection) (d gmac.FileData, err error) {

	d.Separator = "\x00"

	if f.OnlyDirectories {
		d.Operation = "chooseFolder"
	} else {
		d.Operation = "chooseFile"
		d.Opts.OfType = initFiltersOSAScript(f.FileFilters)
	}

	d.Opts.ShowPackages = true
	d.Opts.WithPrompt = f.Title
	d.Opts.Invisibles = f.ShowHidden
	d.Opts.DefaultName = f.Filename
	d.Opts.DefaultLocation = f.Root
	d.Opts.Multiple = f.MultipleSelection

	return
}

func translateFileSave(f *sets.FileSave) (d gmac.FileData, err error) {

	d.Separator = "\x00"

	if f.OnlyDirectories {
		d.Operation = "chooseFolder"
	} else {
		d.Operation = "chooseFileName"
		d.Opts.OfType = initFiltersOSAScript(f.FileFilters)
	}

	d.Opts.ShowPackages = true
	d.Opts.WithPrompt = f.Title
	d.Opts.Invisibles = f.ShowHidden
	d.Opts.DefaultName = f.Filename
	d.Opts.DefaultLocation = f.Root
	d.Opts.Multiple = false

	return
}

func ShowFileSave(ctx context.Context, s *sets.FileSave) (res.File, error) {

	data, err := translateFileSave(s)
	if err != nil {
		return res.File{}, err
	}

	out, err := gmac.RunOSAScript(ctx, "file", data, "")

	if err == nil {
		// Any linebreak at the end will be trimmed away.
		s := strings.TrimSuffix(string(out), "\n")

		return res.File{
			General: res.OkResult(),
			Paths:   strings.Split(s, "\x00")}, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == gmac.ExitCodeCancel {
			return res.File{General: res.CancelResult()}, nil
		}
	}

	return res.File{}, err
}

func ShowFileSelection(ctx context.Context, s *sets.FileSelection) (res.File, error) {

	data, err := translateFileSelection(s)
	if err != nil {
		return res.File{}, err
	}

	out, err := gmac.RunOSAScript(ctx, "file", data, "")

	if err == nil {
		return res.File{
			General: res.OkResult(),
			Paths:   strings.Split(string(out), "\x00")}, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == gmac.ExitCodeCancel {
			return res.File{General: res.CancelResult()}, nil
		}
	}

	return res.File{}, err
}
