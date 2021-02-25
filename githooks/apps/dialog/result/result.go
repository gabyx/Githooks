package result

type resultState = int

const (
	undefinedState resultState = iota
	oKState
	canceledState
	extraButtonState
)

type General struct {
	state          resultState
	extraButtonIdx uint
}

// OkResult creates a accpted res.
func OkResult() General {
	return General{state: oKState}
}

// CancelResult creates a canceled res.
func CancelResult() General {
	return General{state: canceledState}
}

// ExtraButtonResult creates a res.
func ExtraButtonResult(i uint) General {
	return General{state: extraButtonState, extraButtonIdx: i}
}

// IsUnset tells if result is unset.
func (g *General) IsUnset() bool {
	return g.state == undefinedState
}

// IsOk tells if the user clicked ok.
func (g *General) IsOk() bool {
	return g.state == oKState
}

// IsCanceled tells if the user canceled or closed the dialog.
func (g *General) IsCanceled() bool {
	return g.state == canceledState
}

// IsExtraButton tells if the user pressed an extra button.
func (g *General) IsExtraButton() (bool, uint) {
	return g.state == extraButtonState, g.extraButtonIdx
}

// Message is the result type for message dialogs.
type Message struct {
	General
}

// Options is the result type for options dialogs.
type Options struct {
	General

	// The chosen selection indices. Only valid in `IsOk()`.
	Selection []uint
}

// Entry is the result type for options dialogs.
type Entry struct {
	General

	// The entered text.
	Text string
}

// File is the result type for file dialogs.
type File struct {
	General

	// The selected paths.
	Paths []string
}
