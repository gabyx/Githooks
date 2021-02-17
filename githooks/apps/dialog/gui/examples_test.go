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
	t.OkLabel = "Got it"
	t.CancelLabel = "Shut up..." // nolint: goconst
	t.Text = "You have 200 spam mails in your mailbox, you should remove them sooooooon."
	t.Width = 300
	t.Height = 500
	t.WindowIcon = settings.ErrorIcon
	t.Style = settings.ErrorStyle

	_, _ = gui.ShowMessage(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowWarning() { // nolint
	t := settings.Message{}
	t.Title = "SpamMails Alert"
	t.OkLabel = "Got it"
	t.CancelLabel = "Shut up..."
	t.Text = "You have 200 spam mails in your mailbox, you should remove them sooooooon."
	t.Width = 300
	t.Height = 500
	t.WindowIcon = settings.WarningIcon
	t.Style = settings.WarningStyle

	_, _ = gui.ShowMessage(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowQuestion() { // nolint
	t := settings.Message{}
	t.Title = "SpamMails Remove"
	t.OkLabel = "Jeah do it..."
	t.CancelLabel = "Shut up..."
	t.DefaultCancel = true
	t.Text = "You have 200 spam mails in your mailbox, can I remove them?"
	t.Width = 300
	t.Height = 500
	t.WindowIcon = settings.QuestionIcon
	t.Style = settings.QuestionStyle

	_, _ = gui.ShowMessage(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowEntry() {
	t := settings.Entry{}
	t.Title = "SpamMails Remove"
	t.OkLabel = "Jeah accept it..."
	t.CancelLabel = "Ahh cancel..."
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
	t.Title = "Choices" // nolint
	t.OkLabel = "Johh"
	t.CancelLabel = "Nope"
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
	t := settings.FileSave{}
	t.Title = "Choices"
	t.Width = 300
	t.Height = 500
	t.ConfirmOverwrite = true
	t.ConfirmCreate = true
	t.FileFilters = []settings.FileFilter{{Name: "Dev", Patterns: []string{"*.go", "*.sh"}}}
	t.Filename = "MySuperFile/Name.dat"
	t.Root = "../.." // nolint
	t.WindowIcon = settings.QuestionIcon

	f, e := gui.ShowFileSave(nil, &t) // nolint: staticcheck
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
	t.Root = "../.." // nolint
	t.MultipleSelection = true
	t.WindowIcon = settings.QuestionIcon
	t.ShowHidden = false

	f, e := gui.ShowFileSelection(nil, &t) // nolint: staticcheck
	fmt.Fprintf(os.Stderr, "%v, %v", f, e)
	// Output:
}

func ExampleShowDirectorySelection() { // nolint
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
