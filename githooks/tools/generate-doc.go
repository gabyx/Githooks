// +build tools

package main

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"rycus86/githooks/cmd"
	cm "rycus86/githooks/common"
	"rycus86/githooks/git"
	strs "rycus86/githooks/strings"

	"github.com/spf13/cobra/doc"
)

type RegRepl struct {
	Regex *regexp.Regexp
	Repl  string
}

func writeRegexRepl(file string, repls ...RegRepl) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	data := string(bytes)

	for _, r := range repls {
		data = r.Regex.ReplaceAllString(data, r.Repl)
	}

	return ioutil.WriteFile(file, []byte(data), cm.DefaultFileModeFile)

}

func fixFiles(files []string) {

	var regexMDLinks = regexp.MustCompile(`\(git_hooks([a-zA-Z_\-]*\.md)\)`)
	var regexMDLinksBack = regexp.MustCompile(`\(git@hooks([a-zA-Z_\-]*\.md)\)`)
	var regexGithooks = regexp.MustCompile(`git_hooks`)

	const codeVar = `[a-zA-Z0-9_\-:.$+/|{}<> ]`
	const codeVarNS = `[a-zA-Z0-9_\-:.$+/|{}<>]`

	var regexCode = regexp.MustCompile(strs.Fmt(`'(%s+|%s{2,}%s*)'`, codeVarNS, codeVarNS, codeVar))
	var regexListItem = regexp.MustCompile(`•`)

	for _, f := range files {

		err := writeRegexRepl(f,
			RegRepl{Regex: regexMDLinks, Repl: "(git@hooks$1)"},
			RegRepl{Regex: regexGithooks, Repl: "git hooks"},
			RegRepl{Regex: regexCode, Repl: "`$1`"},
			RegRepl{Regex: regexListItem, Repl: "-"},
			RegRepl{Regex: regexMDLinksBack, Repl: "(git_hooks$1)"})

		cm.AssertNoErrorPanic(err, "Replacement failed.")
	}
}

func main() {

	root, err := git.Ctx().Get("rev-parse", "--show-toplevel")
	cm.AssertNoErrorPanic(err, "Could not root dir.")

	docRoot := path.Join(root, "docs")

	log, err := cm.CreateLogContext(false)
	cm.AssertNoErrorPanic(err, "Could not create log")

	ctx := cmd.NewSettings(log)
	cmd := cmd.MakeGithooksCtl(&ctx)
	cmd.Use = "git_hooks" // Fix, because we use a special whitespace...

	err = os.RemoveAll(docRoot)
	cm.AssertNoErrorPanic(err, "Remove failed.")
	err = os.Mkdir(docRoot, cm.DefaultFileModeDirectory)
	cm.AssertNoErrorPanic(err, "Mkdir failed.")

	err = doc.GenMarkdownTree(cmd, docRoot)
	cm.AssertNoErrorPanic(err, "Generating CLI Doc failed.")

	files, err := cm.GetFiles(docRoot, nil)
	cm.AssertNoErrorPanic(err, "Getting files failed.")

	fixFiles(files)
}
