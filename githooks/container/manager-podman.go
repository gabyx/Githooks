package container

import (
	"os/exec"

	cm "github.com/gabyx/githooks/githooks/common"
)

const (
	podmanCmd = "podman"
)

type ManagerPodman struct {
	docker ManagerDocker
}

// ImagePull pulls an image with reference `ref`.
func (m *ManagerPodman) ImagePull(ref string) (err error) {
	return m.docker.ImagePull(ref)
}

// ImageTag tags an image with reference `refSrc` to reference `refTarget`.
func (m *ManagerPodman) ImageTag(refSrc string, refTarget string) (err error) {
	return m.docker.ImageTag(refSrc, refTarget)
}

// ImageBuild builds the stage `stage`
// of an image from `dockerfile` in context path `context` and tags
// it with reference `ref`.
func (m *ManagerPodman) ImageBuild(
	log cm.ILogContext,
	dockerfile string,
	context string,
	stage string,
	ref string) (string, error) {
	return m.docker.ImageBuild(log, dockerfile, context, stage, ref)
}

// ImageExists checks if the image with reference `ref` exists.
func (m *ManagerPodman) ImageExists(ref string) (exists bool, err error) {
	return m.docker.ImageExists(ref)
}

// ImageRemove removes an image with reference `ref`.
func (m *ManagerPodman) ImageRemove(ref string) (err error) {
	return m.docker.ImageRemove(ref)
}

// NewHookRunExec runs a hook over a container.
func (m *ManagerPodman) NewHookRunExec(
	ref string,
	workspaceDir string,
	workspaceHookDir string,
	hookExec cm.IExecutable,
	attachStdIn bool,
	allocateTTY bool,
) (cm.IExecutable, error) {
	return m.docker.NewHookRunExec(ref, workspaceDir, workspaceHookDir, hookExec, attachStdIn, allocateTTY)
}

// IsPodmanAvailable returns if podman is available.
func IsPodmanAvailable() bool {
	_, err := exec.LookPath(podmanCmd)

	return err == nil
}

// NewManagerPodman returns a manger to manage images with podman.
func NewManagerPodman() (IManager, error) {
	if !IsPodmanAvailable() {
		return nil, &ManagerNotAvailableError{podmanCmd}
	}

	return newManagerDocker(podmanCmd, ContainerManagerTypeV.Podman)
}
