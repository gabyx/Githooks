package container

import (
	"os/exec"

	cm "github.com/gabyx/githooks/githooks/common"
)

type ManagerDocker struct {
	cmdCtx cm.CmdContext
}

// ImagePull pulls an image with reference `ref`.
func (m *ManagerDocker) ImagePull(ref string) (err error) {
	return m.cmdCtx.Check("pull", ref)
}

// ImageTag tags an image with reference `refSrc` to reference `refTarget`.
func (m *ManagerDocker) ImageTag(refSrc string, refTarget string) (err error) {
	return m.cmdCtx.Check("tag", refSrc, refTarget)
}

// ImageBuild builds the stage `stage`
// of an image from `dockerfile` in context path `context` and tags
// it with reference `ref`.
func (m *ManagerDocker) ImageBuild(
	log cm.ILogContext,
	dockerfile string,
	context string,
	stage string,
	ref string) (err error) {
	return m.cmdCtx.Check("build", "-f", dockerfile, "-t", ref, "--target", stage, context)
}

// ImageExists checks if the image with reference `ref` exists.
func (m *ManagerDocker) ImageExists(ref string) (exists bool, err error) {
	out, err := m.cmdCtx.GetSplit("image", "ls", "--format", "{{ .ID }}", ref)

	return len(out) != 0, err
}

// ImageRemove removes an image with reference `ref`.
func (m *ManagerDocker) ImageRemove(ref string) (err error) {
	return m.cmdCtx.Check("image", "rm", ref)
}

// IsDockerAvailable returns if docker is available.
func IsDockerAvailable() bool {
	_, err := exec.LookPath("docker")

	return err == nil
}

func NewManagerDocker() (mgr IManager, err error) {
	if !IsDockerAvailable() {
		return nil, &ManagerNotAvailableError{"docker"}
	}

	cmdCtx := cm.NewCommandCtxBuilder().SetBaseCmd("docker").EnableCaptureError().Build()
	mgr = &ManagerDocker{cmdCtx: cmdCtx}

	return
}
