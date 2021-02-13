package common

import (
	set "gabyx/githooks/apps/dialog/settings"

	"github.com/spf13/cobra"
)

func addFlagsGeneral(cmd *cobra.Command, s *set.General) {
	cmd.Flags().StringVar(&s.Title, "title", "", "Dialog title.")
	cmd.Flags().UintVar(&s.Width, "width", 0, "Dialog width.")
	cmd.Flags().UintVar(&s.Height, "height", 0, "Dialog height.")
	cmd.Flags().UintVar((*uint)(&s.WindowIcon), "window-icon", 0, "Window icon.")
}

func addFlagsDefaultButton(cmd *cobra.Command, s *set.DefaultButton) {
	cmd.Flags().StringVar(&s.OkLabel, "ok-label", "", "Ok button label.")
	cmd.Flags().StringVar(&s.CancelLabel, "cancel-label", "", "Cancel button label.")
	cmd.Flags().BoolVar(&s.DefaultCancel, "default-cancel", false, "Set 'Cancel' as the default button.")

	cmd.Flags().StringArrayVar(&s.ExtraButtons, "extra-button", nil, "Extra buttons labels.")
}

func addFlagsGeneralText(cmd *cobra.Command, s *set.GeneralText) {
	cmd.Flags().StringVar(&s.Text, "text", "", "General text of the dialog.")
	cmd.Flags().BoolVar(&s.NoWrap, "no-wrap", false, "Don't wrap the text.")
	cmd.Flags().BoolVar(&s.Ellipsize, "ellipsize", false, "Ellipsize the text if it doesn't fit.")
}

func addFlagsGeneralFile(cmd *cobra.Command, s *set.GeneralFile) {
	cmd.Flags().StringVar(&s.Root, "root", "", "Default root path of the file dialog.")

	cmd.Flags().StringVar(&s.Filename, "filename", "", "Default filename in the dialog.")

	a := fileFilterArgs{Filters: s.FileFilters}
	cmd.Flags().Var(&a, "file-filter", "Sets a filename filter ('<name> | <pattern> <pattern> ...').")

	cmd.Flags().BoolVar(&s.ShowHidden, "show-hidden", false, "Show hidden files.")

	cmd.Flags().BoolVar(&s.OnlyDirectories, "directories", false, "Activate directory-only selection.")
}

func AddFlagsMessage(cmd *cobra.Command, s *set.Message) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralText(cmd, &s.GeneralText)
	addFlagsDefaultButton(cmd, &s.DefaultButton)

	cmd.Flags().UintVar((*uint)(&s.Style), "style", 0, "Message style.")
	cmd.Flags().UintVar((*uint)(&s.Icon), "icon", 0, "Message icon.")
}

func AddFlagsOptions(cmd *cobra.Command, s *set.Options) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralText(cmd, &s.GeneralText)
	addFlagsDefaultButton(cmd, &s.DefaultButton)

	cmd.Flags().StringArrayVar(&s.Options, "option", nil, "Option choice.")
	cmd.Flags().UintVar(&s.DefaultOption, "default-option", 0, "Default option index.")

	cmd.Flags().UintVar((*uint)(&s.Style), "style", 0, "Dialog style: '0' for list, '1' for buttons.")
	cmd.Flags().BoolVar(&s.MultipleSelection, "multiple", false, "Allow multiple selections.")
}

func AddFlagsEntry(cmd *cobra.Command, s *set.Entry) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralText(cmd, &s.GeneralText)
	addFlagsDefaultButton(cmd, &s.DefaultButton)

	cmd.Flags().StringVar(&s.EntryText, "entry-text", "", "The entry text.")
	cmd.Flags().BoolVar(&s.HideEntryText, "hide-text", false, "Hide the text in the entry field.")
}

func AddFlagsNotification(cmd *cobra.Command, s *set.Notification) {
	addFlagsGeneral(cmd, &s.General)
	cmd.Flags().StringVar(&s.Text, "text", "", "Notification text.")
}

func AddFlagsFileSave(cmd *cobra.Command, s *set.FileSave) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralFile(cmd, &s.GeneralFile)

	cmd.Flags().BoolVar(&s.ConfirmOverwrite, "confirm-overwrite", false, "Confirm if the chosen path already exists.")
	cmd.Flags().BoolVar(&s.ShowHidden, "confirm-create", false, "Confirm if the chosen path does not exist.")
}

func AddFlagsFileSelection(cmd *cobra.Command, s *set.FileSelection) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralFile(cmd, &s.GeneralFile)

	cmd.Flags().BoolVar(&s.MultipleSelection, "multiple", false, "Allow multiple selection.")
}
