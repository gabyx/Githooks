package common

import (
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"

	"github.com/spf13/cobra"
)

func addFlagsGeneral(cmd *cobra.Command, s *set.General) {
	cmd.Flags().StringVar(&s.Title, "title", "", "Dialog title.")
	cmd.Flags().UintVar(&s.Width, "width", 0, "Dialog width.")
	cmd.Flags().UintVar(&s.Height, "height", 0, "Dialog height.")

	a := iconArgs{icon: &s.WindowIcon}
	cmd.Flags().Var(&a, "window-icon", `Window icon.
One of ['info', 'warning', 'error', 'question'] (only Windows/Unix)`)
}

func addFlagsGeneralButton(cmd *cobra.Command, s *set.GeneralButton) {
	cmd.Flags().StringVar(&s.OkLabel, "ok-label", "", "Ok button label.")
	cmd.Flags().StringVar(&s.CancelLabel, "cancel-label", "", "Cancel button label.")
	cmd.Flags().BoolVar(&s.DefaultCancel, "default-cancel", false, "Set 'Cancel' as the default button.")

	cmd.Flags().StringArrayVar(&s.ExtraButtons, "extra-button", nil, `Extra buttons labels.
On macOS/Windows only one extra button is allowed.`)
}

func addFlagsGeneralText(cmd *cobra.Command, s *set.GeneralText) {
	cmd.Flags().StringVar(&s.Text, "text", "", "General text of the dialog.")
	cmd.Flags().BoolVar(&s.NoWrap, "no-wrap", false, "Don't wrap the text.")
	cmd.Flags().BoolVar(&s.Ellipsize, "ellipsize", false, "Ellipsize the text if it doesn't fit.")
}

func addFlagsGeneralFile(cmd *cobra.Command, s *set.GeneralFile) {
	cmd.Flags().StringVar(&s.Root, "root", ".", "Default root path of the file dialog.")

	cmd.Flags().StringVar(&s.Filename, "filename", "", "Default filename in the dialog.")

	a := fileFilterArgs{Filters: &s.FileFilters}
	cmd.Flags().Var(&a, "file-filter", "Sets a filename filter ('<name> | <pattern> <pattern> ...').")

	cmd.Flags().BoolVar(&s.ShowHidden, "show-hidden", false, "Show hidden files.")

	cmd.Flags().BoolVar(&s.OnlyDirectories, "directory", false, "Activate directory-only selection.")
}

func AddFlagsMessage(cmd *cobra.Command, s *set.Message) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralText(cmd, &s.GeneralText)
	addFlagsGeneralButton(cmd, &s.GeneralButton)

	a1 := msgStyleArgs{style: &s.Style}
	cmd.Flags().Var(&a1, "style", "Message style.")

	a2 := iconArgs{icon: &s.Icon}
	cmd.Flags().Var(&a2, "icon", `Message icon.
One of ['info', 'warning', 'error', 'question']`)
}

func AddFlagsOptions(cmd *cobra.Command, s *set.Options) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralText(cmd, &s.GeneralText)
	addFlagsGeneralButton(cmd, &s.GeneralButton)

	cmd.Flags().StringArrayVar(&s.Options, "option", nil, "List of options to choose from.")
	a := indexArgs{indices: &s.DefaultOptions}
	cmd.Flags().Var(&a, "default-option", "Default selected option indices (only macOS/Windows).")

	cmd.Flags().UintVar((*uint)(&s.Style), "style", 0,
		`Dialog style: '0' for list, '1' for buttons (only if not '--multiple').
For button style: Default is always either '--default-option' or
the first '--option'.`)
	cmd.Flags().BoolVar(&s.MultipleSelection, "multiple", false, "Allow multiple selections.")
}

func AddFlagsEntry(cmd *cobra.Command, s *set.Entry) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralText(cmd, &s.GeneralText)
	addFlagsGeneralButton(cmd, &s.GeneralButton)

	a2 := iconArgs{icon: &s.Icon}
	cmd.Flags().Var(&a2, "icon", `Message icon (only macOS).
One of ['info', 'warning', 'error', 'question']`)

	cmd.Flags().StringVar(&s.DefaultEntry, "default-entry", "", "The default text in the entry field.")
	cmd.Flags().BoolVar(&s.HideDefaultEntry, "hide-entry", false, "Hide the text in the entry field.")
}

func AddFlagsNotification(cmd *cobra.Command, s *set.Notification) {
	addFlagsGeneral(cmd, &s.General)
	cmd.Flags().StringVar(&s.Text, "text", "", "Notification text.")
}

func AddFlagsFileSave(cmd *cobra.Command, s *set.FileSave) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralFile(cmd, &s.GeneralFile)

	cmd.Flags().BoolVar(&s.ConfirmOverwrite, "confirm-overwrite", false,
		`Confirm if the chosen path already exists.
Cannot be disabled on macOS.`)

	cmd.Flags().BoolVar(&s.ShowHidden, "confirm-create", false,
		"Confirm if the chosen path does not exist (only Windows)")
}

func AddFlagsFileSelection(cmd *cobra.Command, s *set.FileSelection) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralFile(cmd, &s.GeneralFile)

	cmd.Flags().BoolVar(&s.MultipleSelection, "multiple", false, "Allow multiple selection.")
}
