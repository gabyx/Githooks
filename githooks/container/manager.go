package container

import (
	"strings"

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
	Podman ContainerManagerType
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
		workspaceHookDir string,
		exe cm.IExecutable,
		attachStdIn bool,
		allocateTTY bool,
	) (cm.IExecutable, error)
}

// NewManager creates a container manager of type `manager`.
// If empty `docker` is taken.
// Can be a comma-separated string e.g. `podman,docker` to try
// to use the one which first can be constructed.
// Currently only `docker` and `podman` is supported.
func NewManager(manager string) (mgr IManager, err error) {

	if strs.IsEmpty(manager) {
		manager = "docker"
	}

	mgrs := strings.Split(manager, ",")

	var e error
	for _, manager := range mgrs {
		switch manager {
		case "docker":
			mgr, e = NewManagerDocker()
		case "podman":
			mgr, e = NewManagerPodman()
		default:
			e = cm.ErrorF("Container manager '%s' not supported.", manager)
		}

		// If we could construct it, immediately return it.
		if e == nil {
			return mgr, nil
		}

		err = cm.CombineErrors(err, e)
	}

	return nil, cm.CombineErrors(err, cm.Error("Container manager could not be validated."))

}
