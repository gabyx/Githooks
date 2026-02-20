package entry

import (
	"context"
	"os"
	"time"

	dcm "github.com/gabyx/githooks/githooks/apps/dialog/cmd/common"
	"github.com/gabyx/githooks/githooks/apps/dialog/gui"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"

	"github.com/spf13/cobra"
)

func handleResult(ctx *dcm.CmdContext, r *res.Entry, err error) error {
	if ctx.ReportAsJSON {
		return dcm.HandleJSONResult(ctx, res.NewJSONResult(r), &r.General, err)
	}

	return dcm.HandleGeneralResult(ctx, &r.General, err,
		func() error {
			_, e := os.Stdout.WriteString(r.Text + dcm.LineBreak)

			return e
		}, nil, dcm.DefaultExtraButtonCallback(&r.General))
}

// NewCmd creates the entry command.
func NewCmd(ctx *dcm.CmdContext) *cobra.Command {
	settings := set.Entry{}
	var timeout uint

	cmd := &cobra.Command{
		Use:   "entry",
		Short: "Shows a entry dialog.",
		Long: `Shows a entry dialog similar to 'zenity'.
Currently extra buttons are not supported on all platforms.
Unix/Windows supports multiple extra buttons, MacOS does not.

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

			res, err := gui.ShowEntry(cont, &settings)
			err = handleResult(ctx, &res, err)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	cmd.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")

	dcm.AddFlagsEntry(cmd, &settings)
	ccm.SetCommandDefaults(ctx.Log, cmd)

	return cmd
}
