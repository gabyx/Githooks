package result

type actionState = int

const (
	undefinedAction actionState = iota
	oKAction
	canceledAction
	extraButtonAction
)

type General struct {
	Action         actionState
	ExtraButtonIdx uint `yaml:"extraButtonIdx"`
}

// OkResult creates a accpted res.
func OkResult() General {
	return General{Action: oKAction}
}

// CancelResult creates a canceled res.
func CancelResult() General {
	return General{Action: canceledAction}
}

// ExtraButtonResult creates a res.
func ExtraButtonResult(i uint) General {
	return General{Action: extraButtonAction, ExtraButtonIdx: i}
}

// IsUnset tells if result is unset.
func (g *General) IsUnset() bool {
	return g.Action == undefinedAction
}

// IsOk tells if the user clicked ok.
func (g *General) IsOk() bool {
	return g.Action == oKAction
}

// IsCanceled tells if the user canceled or closed the dialog.
func (g *General) IsCanceled() bool {
	return g.Action == canceledAction
}

// IsExtraButton tells if the user pressed an extra button.
func (g *General) IsExtraButton() (bool, uint) {
	return g.Action == extraButtonAction, g.ExtraButtonIdx
}

// Message is the result type for message dialogs.
type Message struct {
	General `yaml:",inline"`
}

// Options is the result type for options dialogs.
type Options struct {
	General `yaml:",inline"`

	// The chosen options indices. Only valid in `IsOk()`.
	Options []uint
}

// Entry is the result type for options dialogs.
type Entry struct {
	General `yaml:",inline"`

	// The entered text.
	Text string
}

// File is the result type for file dialogs.
type File struct {
	General `yaml:",inline"`

	// The selected paths.
	Paths []string
}
