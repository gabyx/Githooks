package gui

type MsgOptions struct {
	Message       string   `json:"message,omitempty"`
	As            string   `json:"as,omitempty"`
	WithTitle     string   `json:"withTitle,omitempty"`
	Buttons       []string `json:"buttons,omitempty"`
	CancelButton  int      `json:"cancelButton,omitempty"`
	DefaultButton int      `json:"defaultButton,omitempty"`

	// For entry only...
	DefaultAnswer string `json:"defaultAnswer"`
	HiddenAnswer  bool   `json:"hiddenAnswer,omitempty"`
}

type MsgData struct {
	Operation string
	Text      string
	WithIcon  string

	Opts MsgOptions `json:"opts"`
}

type OptionsOpts struct {
	WithTitle                string   `json:"withTitle,omitempty"`
	WithPrompt               string   `json:"withPrompt,omitempty"`
	DefaultItems             []string `json:"defaultItems,omitempty"`
	OkButtonName             string   `json:"okButtonName,omitempty"`
	CancelButtonName         string   `json:"cancelButtonName,omitempty"`
	MultipleSelectionAllowed bool     `json:"multipleSelectionsAllowed,omitempty"`
	EmptySelectionAllowed    bool     `json:"emptySelectionAllowed,omitempty"`
}

type OptionsData struct {
	Operation string
	Separator string
	Items     []string

	Opts OptionsOpts
}

type FileOpts struct {
	WithPrompt      string   `json:"withPrompt,omitempty"`
	OfType          []string `json:"ofType,omitempty"`
	DefaultName     string   `json:"defaultName,omitempty"`
	DefaultLocation string   `json:"defaultLocation,omitempty"`
	Invisibles      bool     `json:"invisibles,omitempty"`
	Multiple        bool     `json:"multipleSelectionsAllowed,omitempty"`
	ShowPackages    bool     `json:"showPackageContents,omitempty"`
}

type FileData struct {
	Operation string
	Separator string

	Opts FileOpts
}

type NotifyOpts struct {
	WithTitle string `json:"withTitle"`
	Subtitle  string `json:"subtitle,omitempty"`
}

type NotifyData struct {
	Text string
	Opts NotifyOpts
}
