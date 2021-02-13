package common

import (
	res "gabyx/githooks/apps/dialog/result"
	strs "gabyx/githooks/strings"
	"os"
	"strings"
)

func indicesToList(l []uint) (s []string) {
	for i := range l {
		s = append(s, strs.Fmt("%d", l[i]))
	}

	return
}

// OutputArray outputs a string array to std output.
func OutputArray(l []string, sep string) (err error) {
	if len(l) > 0 {
		_, err = os.Stdout.WriteString(strings.Join(l, sep) + LineBreak)
	}

	return
}

// OutputArray outputs an index array to std output.
func OutputIndexArray(l []uint) error {
	return OutputArray(indicesToList(l), ",")
}

func HandleGeneralResult(ctx *CmdContext,
	g *res.General,
	err error,
	okCallback func() error,
	cancelCallback func() error) error {

	// Handle expected errors first.
	if os.IsTimeout(err) {
		ctx.ExitCode = 5

		return nil

	} else if err != nil {
		// All other errors are not handled.
		return err
	}

	// Handle non-errors.
	if g.IsOk() {
		ctx.ExitCode = 0
		if okCallback != nil {
			e := okCallback()
			if e != nil {
				return e // callback error...
			}
		}
	} else if g.IsCanceled() {
		ctx.ExitCode = 1
		if cancelCallback != nil {
			e := cancelCallback()
			if e != nil {
				return e // callback error...
			}
		}
	} else if clicked, idx := g.IsExtraButton(); clicked {
		os.Stdout.WriteString(strs.Fmt("%d", idx))
		os.Stdout.WriteString(LineBreak)
		ctx.ExitCode = 2
	}

	return nil
}
