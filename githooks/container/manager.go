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

// ContainerMgr provides the interface to `docker` or `podman`
// for the functionality used in Githooks.
// We do not use moby/moby because we would need to wrap agnostic arguments.
type IManager interface {
	ImagePull(ref string) error
	ImageTag(refSrc string, refTarget string) error
	ImageBuild(log cm.ILogContext, dockerfile string, context string, stage string, ref string) error
	ImageExists(ref string) (bool, error)
	ImageRemove(ref string) error
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
