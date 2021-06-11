package updates

import (
	"fmt"
	"strconv"
	"time"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// RecordUpdateCheckTimestamp records the current update check time.
func RecordUpdateCheckTimestamp() error {
	return git.Ctx().SetConfig(hooks.GitCKAutoUpdateCheckTimestamp,
		fmt.Sprintf("%v", time.Now().Unix()), git.GlobalScope)
}

// ResetUpdateCheckTimestamp resets the update check time.
func ResetUpdateCheckTimestamp() error {
	return git.Ctx().UnsetConfig(hooks.GitCKAutoUpdateCheckTimestamp, git.GlobalScope)
}

// GetUpdateCheckTimestamp gets the update check time.
func GetUpdateCheckTimestamp() (t time.Time, isSet bool, err error) {

	// Initialize with too old time...
	t = time.Unix(0, 0)

	timeLastUpdateCheck := git.Ctx().GetConfig(hooks.GitCKAutoUpdateCheckTimestamp, git.GlobalScope)
	if strs.IsEmpty(timeLastUpdateCheck) {
		return
	}
	isSet = true

	value, err := strconv.ParseInt(timeLastUpdateCheck, 10, 64)
	if err != nil {
		err = cm.CombineErrors(cm.Error("Could not parse update time."), err)

		return
	}

	t = time.Unix(value, 0)

	return
}
