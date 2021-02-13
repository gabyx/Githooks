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
	cmd.Flags().StringVar(&s.Text, "text", "", "The general text of the dialog.")
	cmd.Flags().BoolVar(&s.NoWrap, "no-wrap", false, "Don't wrap the text.")
	cmd.Flags().BoolVar(&s.Ellipsize, "ellipsize", false, "Ellipsize the text if it doesn't fit.")
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

	cmd.Flags().StringArrayVar(&s.Options, "option", nil, "Option choices.")
	cmd.Flags().UintVar(&s.DefaultOption, "default-option", 0, "Option choices.")

	cmd.Flags().UintVar((*uint)(&s.Style), "style", 0, "Options style.")
	cmd.Flags().BoolVar(&s.MultipleSelection, "multiple-selections", false, "Multiple selections allowed.")
}

func AddFlagsEntry(cmd *cobra.Command, s *set.Entry) {
	addFlagsGeneral(cmd, &s.General)
	addFlagsGeneralText(cmd, &s.GeneralText)
	addFlagsDefaultButton(cmd, &s.DefaultButton)

	cmd.Flags().StringVar(&s.EntryText, "entry-text", "", "Entry text.")
	cmd.Flags().BoolVar(&s.HideEntryText, "hide-text", false, "Hide the text in the entry field.")
}

func AddFlagsNotification(cmd *cobra.Command, s *set.Notification) {
	addFlagsGeneral(cmd, &s.General)
	cmd.Flags().StringVar(&s.Text, "text", "", "Notification text.")

}
