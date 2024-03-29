package settings

// DialogIcon is the enumeration for dialog icons.
type DialogIcon uint

// The stock dialog icons.
const (
	UndefinedIcon DialogIcon = iota
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

// GeneralButton are default settings for the standard buttons on certain dialogs.
type GeneralButton struct {
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
	InfoStyle MessageStyle = iota
	WarningStyle
	ErrorStyle
	QuestionStyle
)

// Message are options for the message dialog.
type Message struct {
	General
	GeneralText
	GeneralButton

	Style MessageStyle
	Icon  DialogIcon

	ForceTopMost bool // Forces the window to be always on top. Only on Windows supported.
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
	GeneralButton

	Options        []string
	DefaultOptions []uint

	Style             OptionsStyle
	MultipleSelection bool

	ForceTopMost bool // Forces the window to be always on top. Only on Windows supported.
}

// Entry are options for the entry dialog.
type Entry struct {
	General
	GeneralText
	GeneralButton

	DefaultEntry     string
	HideDefaultEntry bool

	Icon DialogIcon // Only on macOS supported.

	ForceTopMost bool // Forces the window to be always on top. Only on Windows supported.
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
// Note: Adding member -> Check Conversion from `FileSave` -> `FileSelection`.
type FileSelection struct {
	General
	GeneralFile

	MultipleSelection bool
}

// SetDefaultIcons sets default icons for message dialogs.
func (s *Message) SetDefaultIcons() {

	switch s.Style {
	case QuestionStyle:
		if s.WindowIcon == UndefinedIcon {
			s.WindowIcon = QuestionIcon
		}
		if s.Icon == UndefinedIcon {
			s.Icon = QuestionIcon
		}
	case InfoStyle:
		if s.WindowIcon == UndefinedIcon {
			s.WindowIcon = InfoIcon
		}
		if s.Icon == UndefinedIcon {
			s.Icon = InfoIcon
		}
	case WarningStyle:
		if s.WindowIcon == UndefinedIcon {
			s.WindowIcon = WarningIcon
		}
		if s.Icon == UndefinedIcon {
			s.Icon = WarningIcon
		}
	case ErrorStyle:
		if s.WindowIcon == UndefinedIcon {
			s.WindowIcon = ErrorIcon
		}
		if s.Icon == UndefinedIcon {
			s.Icon = ErrorIcon
		}
	}
}
