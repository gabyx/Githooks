package gui

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
type GeneralSettings struct {
	Title  string
	Width  uint
	Height uint

	WindowIcon DialogIcon
}

// DefaultButtonSettings are default settings for the standard buttons on certain dialogs.
type DefaultButtonSettings struct {
	OkLabel       string
	CancelLabel   string
	DefaultCancel bool

	ExtraButtons []string
}

// GeneralTextSettings are default text settings for certain dialogs.
type GeneralTextSettings struct {
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

// MessageSettings are options for the message dialog.
type MessageSettings struct {
	GeneralSettings
	GeneralTextSettings
	DefaultButtonSettings

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

// OptionsSettings are options for the options dialog.
type OptionsSettings struct {
	GeneralSettings
	GeneralTextSettings
	DefaultButtonSettings

	Options       []string
	DefaultOption uint

	Style             OptionsStyle
	MultipleSelection bool
}

// EntrySettings are options for the entry dialog.
type EntrySettings struct {
	GeneralSettings
	GeneralTextSettings
	DefaultButtonSettings

	EntryText string
	HideEntry bool
}
