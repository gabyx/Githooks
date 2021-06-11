package message

import (
	"context"
	"time"

	dcm "github.com/gabyx/githooks/githooks/apps/dialog/cmd/common"
	"github.com/gabyx/githooks/githooks/apps/dialog/gui"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"

	"github.com/spf13/cobra"
)

func handleResult(ctx *dcm.CmdContext, r *res.Message, err error) error {
	if ctx.ReportAsJSON {
		return dcm.HandleJSONResult(ctx, res.NewJSONResult(r), &r.General, err)
	}

	return dcm.HandleGeneralResult(
		ctx, &r.General, err,
		nil, nil, dcm.DefaultExtraButtonCallback(&r.General))
}

func NewCmd(ctx *dcm.CmdContext) *cobra.Command {

	settings := set.Message{}
	var timeout uint

	cmd := &cobra.Command{
		Use:   "message",
		Short: "Shows a message dialog.",
		Long: `Shows a message dialog similar to 'zenity'.

Currently only one extra button is supported on all platforms.
Only Unix supports multiple extra buttons.
Use 'options' to have more choices.

# Exit Codes:

- '0' : User pressed 'Ok'.
- '1' : User pressed 'Cancel' or closed the dialog.
- '2' : The user pressed an extra button.
		The output contains the index of that button.`,
		Run: func(cmd *cobra.Command, args []string) {

			var cancel func()
			var cont context.Context

			if timeout > 0 {
				cont, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				defer cancel()
			}

			res, err := gui.ShowMessage(cont, &settings)
			err = handleResult(ctx, &res, err)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	cmd.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")

	dcm.AddFlagsMessage(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
