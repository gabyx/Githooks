package entry

import (
	"context"
	dcm "gabyx/githooks/apps/dialog/cmd/common"
	"gabyx/githooks/apps/dialog/gui"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	ccm "gabyx/githooks/cmd/common"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func handleResult(ctx *dcm.CmdContext, res *res.Entry, err error) error {
	return dcm.HandleGeneralResult(ctx, &res.General, err, func() error {
		_, err := os.Stdout.WriteString(res.Text + dcm.LineBreak)

		return err
	}, nil)
}

func NewCmd(ctx *dcm.CmdContext) *cobra.Command {

	settings := set.Entry{}
	var timeout uint

	cmd := &cobra.Command{
		Use:   "entry",
		Short: "Shows a entry dialog.",
		Long: `Shows a entry dialog similar to 'zenity'.
See 'https://help.gnome.org/users/zenity/3.32' for details.`,
		Run: func(cmd *cobra.Command, args []string) {

			var cancel func()
			var cont context.Context

			if timeout > 0 {
				cont, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				defer cancel()
			}

			res, err := gui.ShowEntry(cont, &settings)
			err = handleResult(ctx, &res, err)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	cmd.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")

	dcm.AddFlagsEntry(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
