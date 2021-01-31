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

	readme := path.Join(root, ".githooks", "README.md")
	template := path.Join(root, "githooks", "run-wrapper.sh")
	deployPGP := path.Join(root, "githooks", ".deploy-pgp")

	return []bindata.InputConfig{
		{Path: template, Recursive: false},
		{Path: readme, Recursive: false},
		{Path: deployPGP, Recursive: false}}
}

func main() {

	root, err := git.Ctx().Get("rev-parse", "--show-toplevel")
	cm.AssertNoErrorPanicF(err, "Could not root dir.")

	srcRoot := path.Join(root, "githooks")

	c := bindata.Config{
		Input:          getFiles(root),
		Package:        pkg,
		NoMemCopy:      false,
		NoCompress:     false,
		HttpFileSystem: false,
		Debug:          false,
		Prefix:         root,
		Output:         path.Join(srcRoot, embeddedFile)}

	err = bindata.Translate(&c)

	cm.AssertNoErrorPanicF(err,
		"Translating files into embedded binary failed.")
}
