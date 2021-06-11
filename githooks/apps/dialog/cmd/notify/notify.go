package notify

import (
	dcm "github.com/gabyx/githooks/githooks/apps/dialog/cmd/common"
	"github.com/gabyx/githooks/githooks/apps/dialog/gui"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"

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

	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Shows a notification.",
		Long: `Shows a notification similar to 'zenity'.

# Exit Codes:

- '0' : The notification was successful.
- > '0' : An error occurred.`,

		Run: func(cmd *cobra.Command, args []string) {
			err := gui.ShowNotification(nil, &settings) //nolint: staticcheck
			err = handleResult(ctx, err)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	dcm.AddFlagsNotification(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
