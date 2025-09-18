package common

import (
	"os"
	"strings"

	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/goccy/go-yaml"
)

func indicesToList(l []uint) (s []string) {
	for i := range l {
		s = append(s, strs.Fmt("%d", l[i]))
	}

	return
}

// OutputArray outputs a string array to std output with an appended line break.
func OutputArray(l []string, sep string) (err error) {
	if len(l) > 0 {
		_, err = os.Stdout.WriteString(strings.Join(l, sep))
	}

	return
}

// OutputString outputs a single string.
func OutputString(s string) (err error) {
	_, err = os.Stdout.WriteString(s)

	return
}

// OutputIndexArray outputs an index array to std output.
func OutputIndexArray(l []uint, sep string) error {
	return OutputArray(indicesToList(l), sep)
}

// DefaultExtraButtonCallback is the default extra button callback.
func DefaultExtraButtonCallback(g *res.General) func() error {
	return func() error {
		_, _ = os.Stdout.WriteString(strs.Fmt("%d%s", g.ExtraButtonIdx, LineBreak))

		return nil
	}
}

// HandleGeneralResult handles output of general result on stdout.
func HandleGeneralResult(ctx *CmdContext,
	g *res.General,
	err error,
	okCallback func() error,
	cancelCallback func() error,
	extraCallback func() error) error {

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
	} else if clicked, _ := g.IsExtraButton(); clicked {
		ctx.ExitCode = 2
		if extraCallback != nil {
			if e := extraCallback(); e != nil {
				return e // callback error...
			}
		}
	}

	return nil
}

// HandleGeneralJSONResult handles the output of the general result.
func HandleGeneralJSONResult(ctx *CmdContext, err error) error {

	// Handle expected errors first.
	if os.IsTimeout(err) {
		return nil
	} else if err != nil {
		// All other errors are not handled.
		return err
	}

	return nil
}

// HandleJSONResult handles the output of the result in form of JSON.
func HandleJSONResult(ctx *CmdContext, r res.JSONResult, g *res.General, err error) error {

	// Handle expected errors first.
	if os.IsTimeout(err) {
		r.Timeout = true
		err = nil
	}

	bytes, e := yaml.Marshal(&r)
	if e != nil {
		return cm.CombineErrors(e, cm.ErrorF("Could not YAML marshal result."))
	}

	bytes, e = yaml.YAMLToJSON(bytes)

	if e != nil {
		return cm.CombineErrors(e, cm.ErrorF("Could not convert to JSON."))
	}

	_, _ = os.Stdout.Write(bytes)

	if err != nil {
		// All other errors are not handled.
		return err
	}

	return nil
}
