// nolint: goconst,govet
package gui_test

import (
	"fmt"
	"os"
	"runtime"

	"github.com/gabyx/githooks/githooks/apps/dialog/gui"
	"github.com/gabyx/githooks/githooks/apps/dialog/settings"
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

	t.Icon = settings.WarningIcon
	_, _ = gui.ShowMessage(nil, &t) // nolint: staticcheck

	t.Icon = settings.InfoIcon
	_, _ = gui.ShowMessage(nil, &t) // nolint: staticcheck

	// Output:
}

func ExampleShowMessageExtra() {
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
	if runtime.GOOS != "darwin" {
		t.ExtraButtons = []string{"Midnight"}
	}
	t.Text = "Enter the time:"
	t.DefaultEntry = "10:30"
	t.Width = 300
	t.Height = 500
	t.Icon = settings.InfoIcon
	t.WindowIcon = settings.InfoIcon

	_, _ = gui.ShowEntry(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowOptions() {
	t := settings.Options{}
	t.Title = "What pizza do you want to order?"
	t.OkLabel = "Okey"
	t.CancelLabel = "Cancel it"
	t.Text = "Choose a pizza from below:"
	t.Options = []string{"Margherita", "Curry & Banana", "Hawai-Style"}
	t.MultipleSelection = true
	t.Width = 300
	t.Height = 500
	t.WindowIcon = settings.QuestionIcon

	_, _ = gui.ShowOptions(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowOptionsButtons() {
	t := settings.Options{}
	t.Title = "What pizza do you want to order?"
	t.OkLabel = "Okey"
	t.CancelLabel = "Cancel it"
	t.Text = "Choose a pizza from below:"
	t.Options = []string{"Margherita", "Curry & Banana", "Hawai-Style"}
	t.MultipleSelection = false
	t.Width = 300
	t.Height = 500
	t.Style = settings.OptionsStyleButtons
	t.WindowIcon = settings.QuestionIcon

	_, _ = gui.ShowOptions(nil, &t) // nolint: staticcheck
	// Output:
}

func ExampleShowFileSave() {
	t := &settings.FileSave{}
	t.Title = "Save the pizza to your desktop:"
	t.Width = 300
	t.Height = 500
	t.ConfirmOverwrite = true
	t.ConfirmCreate = true
	t.FileFilters = []settings.FileFilter{{Name: "Dev", Patterns: []string{"*.go", "*.sh"}}}
	t.Filename = "MySuperFile/Pizza.dat"
	t.Root = "../.."
	t.OnlyDirectories = false
	t.WindowIcon = settings.QuestionIcon

	f, e := gui.ShowFileSave(nil, t) // nolint: staticcheck
	fmt.Fprintf(os.Stderr, "%v, %v", f, e)
	// Output:
}

func ExampleShowFileSelection() {
	t := settings.FileSelection{}
	t.Title = "Select your pizzas:"
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
	t.Title = "Choose your pizza folders:"
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
	t.Title = "Hura: Your pizza has arrived."
	t.Text = "Remove the card box and enjoy your stupid Hawai-Style Pizza.\nWuaaaaaaaa...."
	t.Width = 300
	t.Height = 500
	t.WindowIcon = settings.WarningIcon

	_ = gui.ShowNotification(nil, &t) // nolint: staticcheck
	// Output:
}
