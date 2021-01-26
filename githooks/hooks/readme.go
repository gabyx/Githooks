package hooks

import (
	"gabyx/githooks/build"
	cm "gabyx/githooks/common"
	"os"
	"path"
)

// GetReadmeFile gets the Githooks readme
// file inside a repository hooks directory.
func GetReadmeFile(repoDir string) string {
	return path.Join(GetGithooksDir(repoDir), "README.md")
}

// GetRunWrapperContent gets the bytes of the readme file template.
func getReadmeFileContent() ([]byte, error) {
	return build.Asset(path.Join(HooksDirName, "README.md"))
}

// WriteReadme writes the readme content to `file`.
func WriteReadmeFile(filePath string) (err error) {
	readmeContent, e := getReadmeFileContent()
	cm.AssertNoErrorPanic(e, "Could not get embedded readme content.")

	err = os.MkdirAll(path.Dir(filePath), cm.DefaultFileModeDirectory)
	if err != nil {
		return
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, cm.DefaultFileModeFile)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(readmeContent)
	if err != nil {
		return
	}
	err = file.Sync()
	if err != nil {
		return
	}

	return err
}
