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

	return dcm.HandleGeneralResult(ctx, &res.General, err,
		func() error {
			return cm.CombineErrors(
				dcm.OutputIndexArray(res.Selection, sep),
				dcm.OutputString(last))
		}, nil)
}

func NewCmd(ctx *dcm.CmdContext) *cobra.Command {

	settings := set.Options{}
	var timeout uint
	var separator string

	cmd := &cobra.Command{
		Use:   "options",
		Short: "Shows a options selection dialog.",
		Long: `Shows a list selection dialog similar to 'zenity'.

# Exit Codes:

- '0' : 'Ok' was pressed. The output contains the indices of the selected items
		separated by '--separator'.
- '1' : 'Cancel' was pressed or the dialog was closed.
- '2' : The user pressed an extra button.
		The output contains the index of that button.
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
	cmd.Flags().StringVar(&separator, "separator", ",", "Selection indices separator to use for output, default is ','")

	dcm.AddFlagsOptions(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
