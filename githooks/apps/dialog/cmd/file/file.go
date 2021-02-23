package file

import (
	"context"
	dcm "gabyx/githooks/apps/dialog/cmd/common"
	"gabyx/githooks/apps/dialog/gui"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	ccm "gabyx/githooks/cmd/common"
	cm "gabyx/githooks/common"
	"time"

	"github.com/spf13/cobra"
)

func handleResult(ctx *dcm.CmdContext, res *res.File, err error, sep string) error {

	return dcm.HandleGeneralResult(ctx, &res.General, err, func() error {
		return cm.CombineErrors(
			dcm.OutputArray(res.Paths, sep))
	}, nil, nil)
}

func NewCmd(ctx *dcm.CmdContext) []*cobra.Command {

	setSave := set.FileSave{}
	var timeout uint
	var separator string

	cmdSave := &cobra.Command{
		Use:   "file-save",
		Short: "Shows a file save dialog.",
		Long: `Shows a file save dialog similar to 'zenity'.
# Exit Codes:

- '0' : User pressed 'Ok'. The output contains the selected paths
        separated by '--separator'. All paths use forward slashes
        on any platform.
- '1' : User pressed 'Cancel' or closed the dialog.
- '5' : The dialog was closed due to timeout.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if !cm.IsDirectory(setSave.Root) {
				ctx.Log.PanicF("Root '%s' is not existing.", setSave.Root)
			}
			ccm.PanicIfAnyArgs(ctx.Log)(cmd, args)
		},
		Run: func(cmd *cobra.Command, args []string) {

			var cancel func()
			var cont context.Context

			if timeout > 0 {
				cont, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				defer cancel()
			}

			res, err := gui.ShowFileSave(cont, &setSave)
			err = handleResult(ctx, &res, err, separator)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	cmdSave.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")

	dcm.AddFlagsFileSave(cmdSave, &setSave)
	ccm.SetCommandDefaults(ctx.Log, cmdSave)

	setSelection := set.FileSelection{}
	cmdSelect := &cobra.Command{
		Use:   "file-selection",
		Short: "Shows a file selection dialog.",
		Long: `Shows a file selection dialog similar to 'zenity'.

# Exit Codes:

- '0' : User pressed 'Ok'. The output contains the selected paths
        separated by '--separator'. All paths use forward slashes
        on any platform.
- '1' : User pressed 'Cancel' or closed the dialog.
- '5' : The dialog was closed due to timeout.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if !cm.IsDirectory(setSelection.Root) {
				ctx.Log.PanicF("Root '%s' is not existing.", setSelection.Root)
			}
			ccm.PanicIfAnyArgs(ctx.Log)(cmd, args)
		},
		Run: func(cmd *cobra.Command, args []string) {

			var cancel func()
			var cont context.Context

			if timeout > 0 {
				cont, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
				defer cancel()
			}

			res, err := gui.ShowFileSelection(cont, &setSelection)
			err = handleResult(ctx, &res, err, separator)
			ctx.Log.AssertNoErrorPanic(err, "Dialog failed")
		}}

	cmdSelect.Flags().UintVar(&timeout, "timeout", 0, "Timeout for the dialog")
	cmdSelect.Flags().StringVar(&separator, "separator", "\x00",
		"Path separator to use for output, default is '\x00' (null-terminator).")

	dcm.AddFlagsFileSelection(cmdSelect, &setSelection)
	ccm.SetCommandDefaults(ctx.Log, cmdSelect)

	return []*cobra.Command{cmdSave, cmdSelect}
}
