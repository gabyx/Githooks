//go:build darwin

package gui

import (
	"context"
	"os/exec"
	"strings"

	gmac "github.com/gabyx/githooks/githooks/apps/dialog/gui/darwin"
	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

func translateOptions(opts *sets.Options) (d gmac.OptionsData, err error) {
	d = gmac.OptionsData{}

	d.Operation = "chooseFromList"
	d.Separator = "\x00"
	d.Opts.WithPrompt = opts.Text
	d.Opts.WithTitle = opts.Title

	// Workaround: Append invisible spaces before each item,
	// to identify the index afterwards (no string parsing of the label!).
	d.Items = opts.Options
	id := ""
	for i := range d.Items {
		id += string(idPrefix)
		d.Items[i] = id + d.Items[i]
	}

	d.Opts.OkButtonName = opts.OkLabel
	d.Opts.CancelButtonName = opts.CancelLabel
	d.Opts.MultipleSelectionAllowed = opts.MultipleSelection
	d.Opts.EmptySelectionAllowed = true

	for _, idx := range opts.DefaultOptions {
		if idx >= uint(len(d.Items)) {
			continue
		}
		d.Opts.DefaultItems = append(d.Opts.DefaultItems, d.Items[idx])
	}

	return d, nil
}

func getIndex(item string, maxOptions int) int {

	r := []rune(item)

	i := 0
	for ; i < maxOptions && i < len(r); i++ {
		if r[i] != idPrefix {
			break
		}
	}

	cm.DebugAssert(i > 0)

	return i - 1
}

func getChoicesOSAScript(output string, maxOptions int) (indices []uint) {
	s := strings.TrimSpace(output)
	if strs.IsEmpty(s) {
		return
	}

	out := strings.Split(s, "\x00")
	indices = make([]uint, 0, len(out))

	for i := range out {
		indices = append(indices, uint(getIndex(out[i], maxOptions)))
	}

	return
}

func ShowOptions(ctx context.Context, opts *sets.Options) (r res.Options, err error) {
	zenity, err := gunix.GetZenityExecutable()
	if err == nil {
		return ShowOptionsZenity(ctx, zenity, opts)
	}

	if len(opts.Options) == 0 {
		err = cm.ErrorF("You need at least one option specified.")

		return
	}

	data, err := translateOptions(opts)
	if err != nil {
		return
	}

	out, err := gmac.RunOSAScript(ctx, "options", data, "")

	if err == nil {
		return res.Options{
			General: res.OkResult(),
			Options: getChoicesOSAScript(string(out), len(opts.Options))}, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == gmac.ExitCodeCancel {
			return res.Options{General: res.CancelResult()}, nil
		}
	}

	return
}
