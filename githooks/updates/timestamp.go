package updates

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

var lastUpdateTimeStampFilename = ".last-update-check-timestamp"

func getUpdateCheckTimestampFile(installDir string) string {
	return path.Join(installDir, lastUpdateTimeStampFilename)
}

// RecordUpdateCheckTimestamp records the current update check time.
func RecordUpdateCheckTimestamp(installDir string) error {
	err := os.MkdirAll(installDir, cm.DefaultFileModeDirectory)
	if err != nil {
		return err
	}

	err = os.WriteFile(
		getUpdateCheckTimestampFile(installDir),
		[]byte(fmt.Sprintf("%v", time.Now().Unix())),
		cm.DefaultFileModeFile)

	s, _, _ := GetUpdateCheckTimestamp(installDir)
	fmt.Printf("TIMESTAMPE %s", s)

	return err
}

// ResetUpdateCheckTimestamp resets the update check time.
func ResetUpdateCheckTimestamp(installDir string) error {
	_ = os.Remove(getUpdateCheckTimestampFile(installDir))

	return nil
}

// GetUpdateCheckTimestamp gets the update check time.
func GetUpdateCheckTimestamp(installDir string) (t time.Time, isSet bool, err error) {

	// Initialize with too old time...
	t = time.Unix(0, 0)

	file := getUpdateCheckTimestampFile(installDir)
	timeLastUpdateCheck := ""

	if exists, _ := cm.IsPathExisting(file); exists {
		var data []byte
		data, err = os.ReadFile(getUpdateCheckTimestampFile(installDir))
		if err != nil {
			return
		}

		timeLastUpdateCheck = strings.TrimSpace(string(data))
	} else {
		return
	}

	if strs.IsEmpty(timeLastUpdateCheck) {
		return
	}

	isSet = true
	value, err := strconv.ParseInt(timeLastUpdateCheck, 10, 64) // nolint: gomnd
	if err != nil {
		err = cm.CombineErrors(cm.Error("Could not parse update time."), err)

		return
	}

	t = time.Unix(value, 0)

	return
}
