// nolint: goconst,govet
package gui_test

import (
	"fmt"
	"gabyx/githooks/apps/dialog/gui"
	"gabyx/githooks/apps/dialog/settings"
	"os"
)

func ExampleShowMessage() {
	t := settings.Message{}
	t.Title = "SpamMails Alert"
	t.OkLabel = "Okey"
	t.CancelLabel = "Cancel it"
	t.Text = "You have 200 spam mails in your mailbox, you should remove them sooooooon."
	t.Width = 300
	t.Height = 500
	t.WindowIcon = settings.ErrorIcon
	t.Icon = settings.ErrorIcon
	t.Style = settings.ErrorStyle

	_, _ = gui.ShowMessage(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowWarning() {
	t := settings.Message{}
	t.Title = "SpamMails Alert"
	t.OkLabel = "Okey"
	t.CancelLabel = "Cancel it"
	t.ExtraButtons = []string{"Shut up really..."}
	t.Text = "You have 200 spam mails in your mailbox, you should remove them sooooooon."
	t.Width = 300
	t.Height = 500
	t.Icon = settings.WarningIcon
	t.WindowIcon = settings.WarningIcon
	t.Style = settings.WarningStyle

	_, _ = gui.ShowMessage(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowQuestion() {
	t := settings.Message{}
	t.Title = "SpamMails Remove"
	t.OkLabel = "Okey"
	t.CancelLabel = "Cancel it"
	t.ExtraButtons = []string{"Shut up really..."}
	t.DefaultCancel = true
	t.Text = "You have 200 spam mails in your mailbox, can I remove them?"
	t.Width = 300
	t.Height = 500
	t.Icon = settings.QuestionIcon
	t.WindowIcon = settings.QuestionIcon
	t.Style = settings.QuestionStyle

	_, _ = gui.ShowMessage(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowEntry() {
	t := settings.Entry{}
	t.Title = "SpamMails Remove"
	t.OkLabel = "Accept it"
	t.CancelLabel = "Cancel it"
	t.DefaultCancel = true
	t.Text = "Enter the time:"
	t.EntryText = "10:30"
	t.Width = 300
	t.Height = 500
	t.Icon = settings.InfoIcon
	t.WindowIcon = settings.InfoIcon

	_, _ = gui.ShowEntry(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowOptions() {
	t := settings.Options{}
	t.Title = "Choices"
	t.OkLabel = "Okey"
	t.CancelLabel = "Cancel it"
	t.Text = "Choose some options from below"
	t.Options = []string{"Options 1", "Option 2", "Option 3"}
	t.MultipleSelection = true
	t.Width = 300
	t.Height = 500
	t.WindowIcon = settings.QuestionIcon

	_, _ = gui.ShowOptions(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowFileSave() {
	t := &settings.FileSave{}
	t.Title = "Choices"
	t.Width = 300
	t.Height = 500
	t.ConfirmOverwrite = true
	t.ConfirmCreate = true
	t.FileFilters = []settings.FileFilter{{Name: "Dev", Patterns: []string{"*.go", "*.sh"}}}
	t.Filename = "MySuperFile/Name.dat"
	t.Root = "../.."
	t.OnlyDirectories = true
	t.WindowIcon = settings.QuestionIcon

	f, e := gui.ShowFileSave(nil, t) // nolint: staticcheck
	fmt.Fprintf(os.Stderr, "%v, %v", f, e)
	// Output:
}

func ExampleShowFileSelection() {
	t := settings.FileSelection{}
	t.Title = "Choices"
	t.Width = 300
	t.Height = 500
	t.FileFilters = []settings.FileFilter{{Name: "Dev", Patterns: []string{"*.go", "*.sh"}}}
	t.Filename = "MySuperFile.dat"
	t.Root = "../.."
	t.MultipleSelection = true
	t.WindowIcon = settings.QuestionIcon
	t.ShowHidden = false

	f, e := gui.ShowFileSelection(nil, &t) // nolint: staticcheck
	fmt.Fprintf(os.Stderr, "%v, %v", f, e)
	// Output:
}

func ExampleShowDirectorySelection() {
	t := settings.FileSelection{}
	t.Title = "Choices"
	t.Width = 300
	t.Height = 500
	t.Root = "../.."
	t.MultipleSelection = true
	t.OnlyDirectories = true
	t.WindowIcon = settings.QuestionIcon
	t.ShowHidden = false

	f, e := gui.ShowFileSelection(nil, &t) // nolint: staticcheck
	fmt.Fprintf(os.Stderr, "%v, %v", f, e)
	// Output:
}

func ExampleShowNotification() {
	t := settings.Notification{}
	t.Title = "Wupsi: Lots of spam mail detected."
	t.Text = "Remove your spam mails as soon as possible.\nWuaaaaaaaa...."
	t.Width = 300
	t.Height = 500
	t.WindowIcon = settings.WarningIcon

	_ = gui.ShowNotification(nil, &t) // nolint: staticcheck
	// Output:
}