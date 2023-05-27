package container

import (
	"os/exec"

	cm "github.com/gabyx/githooks/githooks/common"
)

type ManagerDocker struct {
	execCtx cm.IExecContext
}

func (*ManagerDocker) ImagePull(ref string) (err error) {
	return
}
func (*ManagerDocker) ImageTag(refSrc string, refTarget string) (err error) {
	return
}
func (*ManagerDocker) ImageBuild(dockerfile string, context string, target string, ref string) (err error) {
	return
}

func (*ManagerDocker) ImageExists(ref string) (exists bool, err error) {
	return
}

// IsDockerAvailable returns if docker is available.
func IsDockerAvailable() bool {
	_, err := exec.LookPath("docker")

	return err != nil
}

func CreateManagerDocker() (mgr IManager, err error) {
	if !IsDockerAvailable() {
		return nil, &ManagerNotAvailableError{"docker"}
	}

	mgr = &ManagerDocker{execCtx: &cm.ExecContext{}}

	return
}
