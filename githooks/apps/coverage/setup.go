package coverage

import (
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"os"
	"path"
)

// Data holds the counter of the current coverage data file of a run.
// Gets incremented in each run, to accumulate multiple files.
// This is a go-coverage tooling workaround.
type Data struct {
	Counter int `yaml:"counter"`
}

// ReadCoverData reads coverage data.
func ReadCoverData(executableName string) (coverDir string, covDataFile string, covData Data) {
	coverDir = os.Getenv("GH_COVERAGE_DIR")
	cm.PanicIf(strs.IsEmpty(coverDir), "You need to set 'GH_COVERAGE_DIR'")

	if !cm.IsDirectory(coverDir) {
		err := os.MkdirAll(coverDir, cm.DefaultFileModeDirectory)
		cm.AssertNoErrorPanicF(err, "Could not make dir '%s'", coverDir)
	}

	covDataFile = path.Join(coverDir, executableName+".yaml")

	if cm.IsFile(covDataFile) {
		err := cm.LoadYAML(covDataFile, &covData)
		cm.AssertNoErrorPanicF(err, "Could not load '%s'", covDataFile)
	}

	return
}

// Setup setups the coverage stuff.
func Setup(executableName string) (run bool) {

	_, covDataFile, covData := ReadCoverData(executableName)

	// Write the new counter for the next run.
	covData.Counter++
	err := cm.StoreYAML(covDataFile, &covData)
	cm.AssertNoErrorPanicF(err, "Could not store '%s'", covDataFile)

	// Strip flags till...
	for i := range os.Args {
		if os.Args[i] == "githooksCoverage" {
			run = true
			os.Args = append([]string{os.Args[0]}, os.Args[i+1:]...)

			break
		}
	}

	return
}
