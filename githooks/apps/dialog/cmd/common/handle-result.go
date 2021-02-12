package common

import (
	"errors"
	dcm "gabyx/githooks/apps/dialog/common"
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

func HandleOtherErrors(ctx *CmdContext, err error) error {

	if e, ok := err.(*dcm.ErrExtraButton); ok { //nolint: gocritic

		os.Stdout.WriteString(strs.Fmt("%d", e.ButtonIndex))
		os.Stdout.WriteString("\n")
		ctx.ExitCode = 1

		return nil

	} else if errors.Is(err, dcm.ErrCancled) {
		ctx.ExitCode = 1

		return nil

	} else if os.IsTimeout(err) {
		ctx.ExitCode = 5

		return nil

	}

	// All other errors are not handled.
	return err
}

func HandleOutputIndices(ctx *CmdContext, l []uint, err error) error {

	if err == nil {

		if len(l) > 0 {
			os.Stdout.WriteString(strings.Join(indicesToList(l), ","))
			os.Stdout.WriteString("\n")
		}

		return nil
	}

	return HandleOtherErrors(ctx, err)
}
