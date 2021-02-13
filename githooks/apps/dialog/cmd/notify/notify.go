package notify

import (
	"context"
	dcm "gabyx/githooks/apps/dialog/cmd/common"
	"gabyx/githooks/apps/dialog/gui"
	set "gabyx/githooks/apps/dialog/settings"
	ccm "gabyx/githooks/cmd/common"
	"time"

	"github.com/spf13/cobra"
)

func handleResult(ctx *dcm.CmdContext, err error) error {
	if err == nil {
		ctx.ExitCode = 0
	}

	return err
}

func NewCmd(ctx *dcm.CmdContext) *cobra.Command {

	settings := set.Notification{}
	var timeout uint

	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Shows a notification.",
		Long: `Shows a notification similar to 'zenity'.
See 'https://help.gnome.org/users/zenity/3.32' for details.`,
		Run: func(cmd *cobra.Command, args []string) {

			var cancel func()
			var cont context.Context

			if timeout > 0 {
				cont, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				defer cancel()
			}

			err := gui.ShowNotification(cont, &settings)
			err = handleResult(ctx, err)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	cmd.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")

	dcm.AddFlagsNotification(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
