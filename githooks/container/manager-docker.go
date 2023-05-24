package images

import (
	cm "github.com/gabyx/githooks/githooks/common"
)

type ManagerDocker struct {
	execCtx cm.IExecContext
}

func (*ManagerDocker) ImagePull(image string) (err error) {
	return
}
func (*ManagerDocker) ImageTag(imageSrc string, imageTarget string) (err error) {
	return
}
func (*ManagerDocker) ImageBuild(dockerfile string, context string, target string) (err error) {
	return
}

func CreateManagerDocker() (mgr IManager, err error) {
	mgr = &ManagerDocker{execCtx: &cm.ExecContext{}}

	return
}
