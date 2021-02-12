package options

import (
	"context"
	dcm "gabyx/githooks/apps/dialog/cmd/common"
	"gabyx/githooks/apps/dialog/gui"
	ccm "gabyx/githooks/cmd/common"
	"time"

	"github.com/spf13/cobra"
)

func NewCmd(ctx *dcm.CmdContext) *cobra.Command {

	settings := gui.OptionsSettings{}
	var timeout uint

	cmd := &cobra.Command{
		Use:   "options",
		Short: "Shows a options selection dialog.",
		Long: `Shows a list selection dialog similar to 'zenity'.
See 'https://help.gnome.org/users/zenity/3.32/list.html.de' for details.`,
		Run: func(cmd *cobra.Command, args []string) {

			var cancel func()
			var cont context.Context

			if timeout > 0 {
				cont, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				defer cancel()
			}

			res, err := gui.ShowOptions(cont, &settings)

			err = dcm.HandleOutputIndices(ctx, res, err)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")

		}}

	cmd.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")

	dcm.AddFlagsOptionsSettings(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
