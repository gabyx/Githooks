package result

type General struct {
	// The user pressed ok.
	ok bool

	// The user canceled or closed.
	canceledOrClosed bool

	// The user clicked another button.
	extraButton    bool
	extraButtonIdx uint
}

// OkResult creates a accpted res.
func OkResult() General {
	return General{ok: true}
}

// CancelResult creates a canceled res.
func CancelResult() General {
	return General{canceledOrClosed: true}
}

// ExtraButtonResult creates a res.
func ExtraButtonResult(i uint) General {
	return General{extraButton: true, extraButtonIdx: i}
}

// IsUnset tells if result is unset.
func (g *General) IsUnset() bool {
	return !g.ok && !g.canceledOrClosed && !g.extraButton
}

// IsOk tells if the user clicked ok.
func (g *General) IsOk() bool {
	return g.ok && !g.canceledOrClosed
}

// IsCanceled tells if the user canceled or closed the dialog.
func (g *General) IsCanceled() bool {
	return !g.ok && g.canceledOrClosed
}

// IsExtraButton tells if the user pressed an extra button.
func (g *General) IsExtraButton() (bool, uint) {
	return g.extraButton, g.extraButtonIdx
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
