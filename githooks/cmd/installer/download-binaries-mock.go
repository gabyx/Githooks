//go:build mock

package installer

import (
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
	"github.com/gabyx/githooks/githooks/updates"
	"github.com/gabyx/githooks/githooks/updates/download"
	"os"
	"path"
	"runtime"
)

// IsRunningCoverage returns if we are running coverage.
var IsRunningCoverage = strs.IsNotEmpty(os.Getenv("GH_COVERAGE_DIR"))

// detectDeploySettings tries to detect the deploy settings.
// Currently that works for Github automatically.
// For Gitea you need to specify the deploy api `deployAPI`.
// Others will fail and need a special deploy settings config file.
func detectDeploySettings(cloneUrl string, deployAPI string) (download.IDeploySettings, error) {
	return nil, nil
}

func downloadBinaries(
	log cm.ILogContext,
	deploySettings download.IDeploySettings,
	tempDir string,
	versionTag string) updates.Binaries {

	bin := os.Getenv("GH_TEST_BIN")
	cm.PanicIf(strs.IsEmpty(bin), "GH_TEST_BIN undefined")

	log.InfoF("Faking download: taking from '%s'.", bin)

	ext := ""
	if runtime.GOOS == cm.WindowsOsName {
		ext = cm.WindowsExecutableSuffix
	}

	all := []string{
		path.Join(tempDir, "cli"+ext),
		path.Join(tempDir, "runner"+ext),
		path.Join(tempDir, "dialog"+ext)}

	for _, exe := range all {
		src := path.Join(bin, path.Base(exe))
		err := cm.CopyFileOrDirectory(src, exe)
		cm.AssertNoErrorPanicF(err, "Copy from '%s' to '%s' failed.", src, exe)
	}

	return updates.Binaries{All: all, Cli: all[0], Others: all[1:]}
}
