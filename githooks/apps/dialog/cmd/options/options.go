package options

import (
	"context"
	"time"

	dcm "github.com/gabyx/githooks/githooks/apps/dialog/cmd/common"
	"github.com/gabyx/githooks/githooks/apps/dialog/gui"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/spf13/cobra"
)

func handleResult(ctx *dcm.CmdContext, r *res.Options, err error, sep string) error {

	if ctx.ReportAsJSON {
		return dcm.HandleJSONResult(ctx, res.NewJSONResult(r), &r.General, err)
	}

	last := ""
	if strs.IsEmpty(sep) {
		// If a separator is chosen don't output last linebreak.
		// this for 'xargs' compatibility.
		sep = ","
		last = dcm.LineBreak
	}

	printRes := func() error {
		return cm.CombineErrors(
			dcm.OutputIndexArray(r.Options, sep),
			dcm.OutputString(last))
	}

	return dcm.HandleGeneralResult(ctx, &r.General, err,
		printRes, nil, dcm.DefaultExtraButtonCallback(&r.General))
}

// NewCmd creates the options command.
func NewCmd(ctx *dcm.CmdContext) *cobra.Command {

	settings := set.Options{}
	var timeout uint
	var separator string

	cmd := &cobra.Command{
		Use:   "options",
		Short: "Shows a options selection dialog.",
		Long: `Shows a list selection dialog similar to 'zenity'.

Extra buttons are only supported on Unix and Windows.
If not using '--multiple' you can also use the
button style options with '--style 1' which uses buttons instead
of a listbox.

# Exit Codes:

- '0' : 'Ok' was pressed. The output contains the indices of the selected items
        separated by '--separator'.
- '1' : 'Cancel' was pressed or the dialog was closed.
- '2' : The user pressed an extra button.
        The output contains the index of that button on the first line.
- '5' : The dialog was closed due to timeout.`,
		Run: func(cmd *cobra.Command, args []string) {

			var cancel func()
			var cont context.Context

			if timeout > 0 {
				cont, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				defer cancel()
			}

			res, err := gui.ShowOptions(cont, &settings)
			err = handleResult(ctx, &res, err, separator)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	cmd.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")
	cmd.Flags().StringVar(&separator,
		"separator",
		"", // empty intentionally
		"Selection indices separator to use for output, default is ','")

	dcm.AddFlagsOptions(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
