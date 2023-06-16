package container

import (
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

type ManagerNotAvailableError struct {
	Cmd string
}

func (m *ManagerNotAvailableError) Error() string {
	return strs.Fmt("Container manager command '%s' not available.", m.Cmd)
}

// EnvVariableContainerRun is the environment variable which is
// set to true in containerized runs.
const EnvVariableContainerRun = "GITHOOKS_CONTAINER_RUN"

type ContainerManagerType int
type containerManagerType struct {
	Docker ContainerManagerType
	Podman ContainerManagerType // Not yet supported.
}

// ContainerManagerTypeV enumerates all container managers supported so far.
var ContainerManagerTypeV = &containerManagerType{Docker: 0, Podman: 1}

// ContainerMgr provides the interface to `docker` or `podman` (etc.)
// for the functionality used in Githooks.
type IManager interface {
	ImagePull(ref string) error
	ImageTag(refSrc string, refTarget string) error
	ImageBuild(
		log cm.ILogContext,
		dockerfile string,
		context string,
		stage string,
		ref string) (string, error)
	ImageExists(ref string) (bool, error)
	ImageRemove(ref string) error

	NewHookRunExec(
		ref string,
		workspaceDir string,
		hookRepoDir string,
		exe cm.IExecutable,
	) (cm.IExecutable, error)
}

// NewManager creates a container manager of type `manager`.
// If empty `docker` is taken.
// Currently only `docker` is supported.
func NewManager(manager string) (mgr IManager, err error) {

	if strs.IsEmpty(manager) {
		manager = "docker"
	}

	switch manager {
	case "docker":
		mgr, err = NewManagerDocker()
	default:
		return nil, cm.ErrorF("Container manager '%s' not supported.", manager)
	}

	return

}
