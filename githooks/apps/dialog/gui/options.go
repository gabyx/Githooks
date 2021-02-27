package gui

import (
	"context"
	res "gabyx/githooks/apps/dialog/result"
	sets "gabyx/githooks/apps/dialog/settings"
)

// showOptionsWithButtons shows a option dialog with buttons instead of a list box.
func showOptionsWithButtons(ctx context.Context, opts *sets.Options) (r res.Options, err error) {

	// Wrap it through `ShowMessage`.
	msg := sets.Message{
		General:       opts.General,
		GeneralText:   opts.GeneralText,
		GeneralButton: opts.GeneralButton,
		Style:         sets.InfoStyle,
		Icon:          opts.WindowIcon}

	// Default configuration
	extraButtons := append([]string{}, opts.Options[1:]...)
	msg.OkLabel = opts.Options[0]
	okOptionIdx := uint(0)
	extraOptionIdx := []uint{1, 2, 3}

	dO := len(opts.DefaultOptions)
	// Swap default configuration with the default option.
	if dO != 0 && opts.DefaultOptions[dO-1] < uint(len(opts.Options)) {
		idx := opts.DefaultOptions[dO-1]
		for i, eIdx := range extraOptionIdx {
			if eIdx == idx {
				// Swap indices...
				extraOptionIdx[i], okOptionIdx = okOptionIdx, extraOptionIdx[i]
				msg.OkLabel, extraButtons[i] = extraButtons[i], msg.OkLabel
			}
		}
	}

	msg.ExtraButtons = append(extraButtons, msg.ExtraButtons...)
	mRes, err := ShowMessage(ctx, &msg)

	if err == nil {
		if mRes.IsOk() {
			return res.Options{
				General: res.OkResult(),
				Options: []uint{okOptionIdx}}, nil

		} else if ok, idx := mRes.IsExtraButton(); ok {

			nSkip := uint(len(opts.Options) - 1)
			if idx < nSkip {
				return res.Options{
					General: res.OkResult(),
					Options: []uint{extraOptionIdx[idx]}}, nil
			}

			return res.Options{General: res.ExtraButtonResult(idx - nSkip)}, nil

		} else if mRes.IsCanceled() {
			return res.Options{General: res.CancelResult()}, nil
		}
	}

	return
}
