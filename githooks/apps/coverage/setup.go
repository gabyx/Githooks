package coverage

import (
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"os"
	"path"
)

type coverageData struct {
	Counter int `yaml:"counter"`
}

func Setup(executableName string) {

	coverDir := os.Getenv("GH_COVERAGE_DIR")

	if strs.IsEmpty(coverDir) {
		cm.Panic("You need to set 'GH_COVERAGE_DIR'")
	} else if !cm.IsDirectory(coverDir) {
		err := os.MkdirAll(coverDir, cm.DefaultFileModeDirectory)
		cm.AssertNoErrorPanicF(err, "Could not make dir '%s'", coverDir)
	}

	covData := coverageData{}
	covDataFile := path.Join(coverDir, executableName+".yaml")

	if cm.IsFile(covDataFile) {
		// Increase the counter for the test files
		err := cm.LoadYAML(covDataFile, &covData)
		cm.AssertNoErrorPanicF(err, "Could not load '%s'", covDataFile)
	}

	// Write the new counter for the next run.
	covData.Counter += 1
	err := cm.StoreYAML(covDataFile, &covData)
	cm.AssertNoErrorPanicF(err, "Could not store '%s'", covDataFile)

	// Strip flags till...
	for i := range os.Args {
		if os.Args[i] == "-githooksCoverage" {
			os.Args = append([]string{os.Args[0]}, os.Args[i+1:]...)

			break
		}
	}
}
