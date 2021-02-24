// +build tools

package main

import (
	"gabyx/githooks/apps/dialog/cmd"
	dcm "gabyx/githooks/apps/dialog/cmd/common"
	cm "gabyx/githooks/common"
	"gabyx/githooks/git"
	"os"
	"path"
	"regexp"

	"github.com/spf13/cobra/doc"
)

type RegRepl struct {
	Regex *regexp.Regexp
	Repl  string
}

func main() {

	root, err := git.Ctx().Get("rev-parse", "--show-toplevel")
	cm.AssertNoErrorPanic(err, "Could not root dir.")

	docRoot := path.Join(root, "docs", "dialog")

	log, err := cm.CreateLogContext(false)
	cm.AssertNoErrorPanic(err, "Could not create log")

	ctx := dcm.CmdContext{Log: log}
	cmd := cmd.MakeDialogCtl(&ctx)

	err = os.RemoveAll(docRoot)
	cm.AssertNoErrorPanic(err, "Remove failed.")
	err = os.Mkdir(docRoot, cm.DefaultFileModeDirectory)
	cm.AssertNoErrorPanic(err, "Mkdir failed.")

	err = doc.GenMarkdownTree(cmd, docRoot)
	cm.AssertNoErrorPanic(err, "Generating CLI Dialog Doc failed.")
}
