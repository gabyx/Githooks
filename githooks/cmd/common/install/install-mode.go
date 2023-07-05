package install

import (
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
)

type InstallModeType int
type installModeType struct {
	None          InstallModeType
	TemplateDir   InstallModeType
	CoreHooksPath InstallModeType
	Manual        InstallModeType
}

// InstallModeTypeV enumerates all types of install modes.
var InstallModeTypeV = &installModeType{TemplateDir: 0, CoreHooksPath: 1, Manual: 2, None: 3} // nolint:gomnd

// GetInstallMode returns the current set install mode of Githooks.
func GetInstallMode(gitx *git.Context) InstallModeType {
	useManual := gitx.GetConfig(hooks.GitCKUseManual, git.GlobalScope) == git.GitCVTrue
	useCoreHooksPathValue := gitx.GetConfig(hooks.GitCKUseCoreHooksPath, git.GlobalScope)

	switch {
	case useManual:
		return InstallModeTypeV.Manual
	case useCoreHooksPathValue == git.GitCVTrue:
		return InstallModeTypeV.CoreHooksPath
	case useCoreHooksPathValue == git.GitCVFalse:
		return InstallModeTypeV.TemplateDir
	default:
		return InstallModeTypeV.None
	}

}

// GetInstallModeName returns a string for the install mode.
func GetInstallModeName(installMode InstallModeType) string {
	switch installMode {
	case InstallModeTypeV.Manual:
		return "manual"
	case InstallModeTypeV.TemplateDir:
		return "template-dir"
	case InstallModeTypeV.CoreHooksPath:
		return "core-hooks-path"
	default:
		return "none"
	}
}

// MapInstallerArgsToInstallMode maps installer arguments to install modes.
func MapInstallerArgsToInstallMode(
	useCoreHooksPath bool,
	useManual bool) InstallModeType {

	switch {
	case useManual:
		return InstallModeTypeV.Manual
	case useCoreHooksPath:
		return InstallModeTypeV.CoreHooksPath
	default:
		return InstallModeTypeV.TemplateDir
	}
}
