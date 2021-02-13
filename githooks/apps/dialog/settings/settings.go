package settings

// DialogIcon is the enumeration for dialog icons.
type DialogIcon uint

// The stock dialog icons.
const (
	Undefined DialogIcon = iota
	ErrorIcon
	WarningIcon
	InfoIcon
	QuestionIcon
)

// General Dialog settings are default settings for all dialogs.
type General struct {
	Title  string
	Width  uint
	Height uint

	WindowIcon DialogIcon
}

// DefaultButton are default settings for the standard buttons on certain dialogs.
type DefaultButton struct {
	OkLabel       string
	CancelLabel   string
	DefaultCancel bool

	ExtraButtons []string
}

// GeneralText are default text settings for certain dialogs.
type GeneralText struct {
	Text      string
	NoWrap    bool
	Ellipsize bool
}

// MessageStyle is the message style.
type MessageStyle uint

// The message styles.
const (
	QuestionStyle MessageStyle = iota
	InfoStyle
	WarningStyle
	ErrorStyle
)

// Message are options for the message dialog.
type Message struct {
	General
	GeneralText
	DefaultButton

	Style MessageStyle
	Icon  DialogIcon
}

// OptionsStyle is the style with which the list dialog is rendered.
type OptionsStyle uint

const (
	// OptionsStyleList renders the list dialog with a selection list.
	OptionsStyleList OptionsStyle = iota

	// OptionsStyleButtons renders the list dialog with a buttons list.
	// Multiple selections can therefore not be performed.
	OptionsStyleButtons
)

// Options are options for the options dialog.
type Options struct {
	General
	GeneralText
	DefaultButton

	Options       []string
	DefaultOption uint

	Style             OptionsStyle
	MultipleSelection bool
}

// Entry are options for the entry dialog.
type Entry struct {
	General
	GeneralText
	DefaultButton

	EntryText     string
	HideEntryText bool
}

// Notification are options for the notification.
type Notification struct {
	General

	Text string
}
