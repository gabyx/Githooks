package gui

type MsgOptions struct {
	Message       string   `json:"message,omitempty"`
	As            string   `json:"as,omitempty"`
	WithTitle     string   `json:"withTitle,omitempty"`
	Buttons       []string `json:"buttons,omitempty"`
	CancelButton  int      `json:"cancelButton,omitempty"`
	DefaultButton int      `json:"defaultButton,omitempty"`
}

type MsgData struct {
	Operation string
	Text      string
	WithIcon  string

	Opts MsgOptions `json:"opts"`
}

type FileOpts struct {
	WithPrompt      string `json:"withPrompt,omitempty"`
	OfType          string `json:"ofType,omitempty"`
	DefaultName     string `json:"defaultName,omitempty"`
	DefaultLocation string `json:"defaultLocation,omitempty"`
	Invisibles      bool   `json:"invisibles,omitempty"`
	Multiple        bool   `json:"multiple,omitempty"`
}

type FileData struct {
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
