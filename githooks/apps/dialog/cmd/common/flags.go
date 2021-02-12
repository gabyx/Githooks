package common

import (
	"gabyx/githooks/apps/dialog/gui"

	"github.com/spf13/cobra"
)

func AddFlagsGeneralSettings(cmd *cobra.Command, s *gui.GeneralSettings) {
	cmd.Flags().StringVar(&s.Title, "title", "", "Dialog title.")
	cmd.Flags().UintVar(&s.Width, "width", 0, "Dialog width.")
	cmd.Flags().UintVar(&s.Height, "height", 0, "Dialog height.")
	cmd.Flags().UintVar((*uint)(&s.WindowIcon), "window-icon", 0, "Window icon.")
}

func AddFlagsDefaultButtonSettings(cmd *cobra.Command, s *gui.DefaultButtonSettings) {
	cmd.Flags().StringVar(&s.OkLabel, "ok-label", "", "Ok button label.")
	cmd.Flags().StringVar(&s.CancelLabel, "cancel-label", "", "Cancel button label.")
	cmd.Flags().BoolVar(&s.DefaultCancel, "default-cancel", false, "Set 'Cancel' as the default button.")

	cmd.Flags().StringArrayVar(&s.ExtraButtons, "extra-button", nil, "Extra buttons labels.")

}

func AddFlagsGeneralTextSettings(cmd *cobra.Command, s *gui.GeneralTextSettings) {
	cmd.Flags().StringVar(&s.Text, "text", "", "The general text of the dialog.")
	cmd.Flags().BoolVar(&s.NoWrap, "no-wrap", false, "Don't wrap the text.")
	cmd.Flags().BoolVar(&s.Ellipsize, "ellipsize", false, "Ellipsize the text if it doesn't fit.")
}

func AddFlagsMessageSettings(cmd *cobra.Command, s *gui.MessageSettings) {
	AddFlagsGeneralSettings(cmd, &s.GeneralSettings)
	AddFlagsGeneralTextSettings(cmd, &s.GeneralTextSettings)
	AddFlagsDefaultButtonSettings(cmd, &s.DefaultButtonSettings)

	cmd.Flags().UintVar((*uint)(&s.Style), "style", 0, "Message style.")
	cmd.Flags().UintVar((*uint)(&s.Icon), "icon", 0, "Message icon.")
}

func AddFlagsOptionsSettings(cmd *cobra.Command, s *gui.OptionsSettings) {
	AddFlagsGeneralSettings(cmd, &s.GeneralSettings)
	AddFlagsGeneralTextSettings(cmd, &s.GeneralTextSettings)
	AddFlagsDefaultButtonSettings(cmd, &s.DefaultButtonSettings)

	cmd.Flags().StringArrayVar(&s.Options, "option", nil, "Option choices.")
	cmd.Flags().UintVar(&s.DefaultOption, "default-option", 0, "Option choices.")

	cmd.Flags().UintVar((*uint)(&s.Style), "style", 0, "Options style.")
	cmd.Flags().BoolVar(&s.MultipleSelection, "multiple-selections", false, "Multiple selections allowed.")
}

func AddFlagsEntrySettings(cmd *cobra.Command, s *gui.EntrySettings) {
	AddFlagsGeneralSettings(cmd, &s.GeneralSettings)
	AddFlagsGeneralTextSettings(cmd, &s.GeneralTextSettings)
	AddFlagsDefaultButtonSettings(cmd, &s.DefaultButtonSettings)

	cmd.Flags().StringVar(&s.EntryText, "--entry-text", "", "Entry text.")
	cmd.Flags().BoolVar(&s.HideEntry, "ellipsize", true, "Ellipsize the text if it doesn't fit.")
}
