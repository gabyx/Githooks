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

	ExtraButtons []string // On macOS only one additional button is allowed.
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
	// Unsupported at the moment.
	OptionsStyleButtons
)

// Options are options for the options dialog.
type Options struct {
	General
	GeneralText
	DefaultButton

	Options        []string
	DefaultOptions []uint

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

	Icon DialogIcon // Only on macOS supported.
}

// Notification are options for the notification.
type Notification struct {
	General

	Text string
}

// FileFilter is file filter for the file dialog.
type FileFilter struct {
	Name     string   // Name that describes the filter (optional).
	Patterns []string // Filter patterns for the display string.
}

// GeneralFile are general options for file dialog.
type GeneralFile struct {
	// Root directory of the file dialog. (default is current working dir)
	Root string

	// Default file/directory name in the dialog.
	Filename string

	// File filter.
	FileFilters []FileFilter

	// Show hidden files.
	ShowHidden bool // Windows and macOS only.

	// Select only directories.
	OnlyDirectories bool
}

// FileSave are options for the file dialog.
type FileSave struct {
	General
	GeneralFile

	// Confirm if the file get overwritten.
	// Cannot be disabled on macOS.
	ConfirmOverwrite bool

	// Confirm if the file does not exist.
	ConfirmCreate bool // Windows only.
}

// FileSelection are options for the file dialog.
type FileSelection struct {
	General
	GeneralFile

	MultipleSelection bool
}
