package coverage

import (
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// Data for coverage tooling.
// A counter which increments in each run such that we can accumulate different coverage reports.
// This is a stupid workaround.
type Data struct {
	Counter int `yaml:"counter"`
}

// ReadCoverData reads data for coverage tooling.
func ReadCoverData(executableName string) (coverDir string, covDataFile string, covData Data) {
	coverDir = os.Getenv("GH_COVERAGE_DIR")

	if strs.IsEmpty(coverDir) {
		cm.Panic("You need to set 'GH_COVERAGE_DIR'")
	} else if !cm.IsDirectory(coverDir) {
		err := os.MkdirAll(coverDir, cm.DefaultFileModeDirectory)
		cm.AssertNoErrorPanicF(err, "Could not make dir '%s'", coverDir)
	}

	covDataFile = path.Join(coverDir, executableName+".yaml")

	if cm.IsFile(covDataFile) {
		// Increase the counter for the test files
		err := cm.LoadYAML(covDataFile, &covData)
		cm.AssertNoErrorPanicF(err, "Could not load '%s'", covDataFile)
	}

	return
}

// Setup setups coverage tooling stuff.
func Setup(executableName string) {

	_, covDataFile, covData := ReadCoverData(executableName)

	// Write the new counter for the next run.
	covData.Counter++
	err := cm.StoreYAML(covDataFile, &covData)
	cm.AssertNoErrorPanicF(err, "Could not store '%s'", covDataFile)

	// Strip flags till...
	for i := range os.Args {
		if os.Args[i] == "githooksCoverage" {
			os.Args = append([]string{os.Args[0]}, os.Args[i+1:]...)

			break
		}
	}
}
