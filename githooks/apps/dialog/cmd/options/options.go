package options

import (
	"context"
	dcm "gabyx/githooks/apps/dialog/cmd/common"
	"gabyx/githooks/apps/dialog/gui"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	ccm "gabyx/githooks/cmd/common"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"time"

	"github.com/spf13/cobra"
)

func handleResult(ctx *dcm.CmdContext, res *res.Options, err error, sep string) error {

	last := ""
	if strs.IsEmpty(sep) {
		// If a separator is chosen don't output last linebreak.
		// this for 'xargs' compatibility.
		sep = ","
		last = dcm.LineBreak
	}

	printRes := func() error {
		return cm.CombineErrors(
			dcm.OutputIndexArray(res.Selection, sep),
			dcm.OutputString(last))
	}

	return dcm.HandleGeneralResult(ctx, &res.General, err,
		printRes, nil, nil)
}

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
