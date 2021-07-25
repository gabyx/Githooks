package gui

// MsgOptions holds all options for a message dialog.
type MsgOptions struct {
	Message       string   `json:"message,omitempty"`
	As            string   `json:"as,omitempty"`
	WithTitle     string   `json:"withTitle,omitempty"`
	Buttons       []string `json:"buttons,omitempty"`
	CancelButton  int      `json:"cancelButton,omitempty"`
	DefaultButton int      `json:"defaultButton,omitempty"`
}

// MsgData holds data for a message dialog.
type MsgData struct {
	Operation string
	Text      string
	WithIcon  string

	Opts MsgOptions `json:"opts"`
}

// EntryOpts holds all options for an entry dialog.
type EntryOpts struct {
	MsgOptions

	DefaultAnswer string `json:"defaultAnswer"`
	HiddenAnswer  bool   `json:"hiddenAnswer,omitempty"`
}

// EntryData holds data for an entry dialog.
// Note: Adding member -> Check `NewFromEntry`.
type EntryData struct {
	Operation string
	Text      string
	WithIcon  string

	Opts EntryOpts `json:"opts"`
}

// NewFromEntry creates new entry data from message data.
func NewFromEntry(m *MsgData) EntryData {
	return EntryData{
		Operation: m.Operation,
		Text:      m.Text,
		WithIcon:  m.WithIcon,
		Opts:      EntryOpts{MsgOptions: m.Opts},
	}
}

// OptionsOpts holds all options for an options dialog.
type OptionsOpts struct {
	WithTitle                string   `json:"withTitle,omitempty"`
	WithPrompt               string   `json:"withPrompt,omitempty"`
	DefaultItems             []string `json:"defaultItems,omitempty"`
	OkButtonName             string   `json:"okButtonName,omitempty"`
	CancelButtonName         string   `json:"cancelButtonName,omitempty"`
	MultipleSelectionAllowed bool     `json:"multipleSelectionsAllowed,omitempty"`
	EmptySelectionAllowed    bool     `json:"emptySelectionAllowed,omitempty"`
}

// OptionsData holds additional options for an options dialog.
type OptionsData struct {
	Operation string
	Separator string
	Items     []string

	Opts OptionsOpts
}

// FileOpts holds all options for a file dialog.
type FileOpts struct {
	WithPrompt      string   `json:"withPrompt,omitempty"`
	OfType          []string `json:"ofType,omitempty"`
	DefaultName     string   `json:"defaultName,omitempty"`
	DefaultLocation string   `json:"defaultLocation,omitempty"`
	Invisibles      bool     `json:"invisibles,omitempty"`
	Multiple        bool     `json:"multipleSelectionsAllowed,omitempty"`
	ShowPackages    bool     `json:"showPackageContents,omitempty"`
}

// FileData holds all data for a file dialog.
type FileData struct {
	Operation string
	Separator string

	Opts FileOpts
}

// NotifyOpts holds all options for a system notification.
type NotifyOpts struct {
	WithTitle string `json:"withTitle"`
	Subtitle  string `json:"subtitle,omitempty"`
}

// NotifyData holds all data for a system notification.
type NotifyData struct {
	Text string
	Opts NotifyOpts
}
