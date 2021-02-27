// +build tools

package main

import (
	cm "gabyx/githooks/common"
	"gabyx/githooks/git"
	"path"

	"github.com/go-bindata/go-bindata"
)

var pkg = "build"
var embeddedFile = "build/embedded-files.go"

func getFiles(root string) []bindata.InputConfig {

	// Embed macOS script files
	filesTmpl := path.Join(root, "gui", "darwin", "osascripts", "file.js.tmpl")
	msgTmpl := path.Join(root, "gui", "darwin", "osascripts", "message.js.tmpl")
	notifyTmpl := path.Join(root, "gui", "darwin", "osascripts", "notify.js.tmpl")
	optionsTmpl := path.Join(root, "gui", "darwin", "osascripts", "options.js.tmpl")
	entryTmpl := path.Join(root, "gui", "darwin", "osascripts", "entry.js.tmpl")

	return []bindata.InputConfig{
		{Path: filesTmpl, Recursive: false},
		{Path: msgTmpl, Recursive: false},
		{Path: notifyTmpl, Recursive: false},
		{Path: optionsTmpl, Recursive: false},
		{Path: entryTmpl, Recursive: false}}
}

func main() {

	root, err := git.Ctx().Get("rev-parse", "--show-toplevel")
	cm.AssertNoErrorPanicF(err, "Could not root dir.")

	srcRoot := path.Join(root, "githooks", "apps", "dialog")

	c := bindata.Config{
		Input:          getFiles(srcRoot),
		Package:        pkg,
		NoMemCopy:      false,
		NoCompress:     false,
		HttpFileSystem: false,
		Debug:          false,
		Prefix:         srcRoot,
		Output:         path.Join(srcRoot, embeddedFile)}

	err = bindata.Translate(&c)

	cm.AssertNoErrorPanicF(err,
		"Translating files into embedded binary failed.")
}
