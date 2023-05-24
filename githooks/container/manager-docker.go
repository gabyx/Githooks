package container

import (
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

func CreateManagerDocker() (mgr IManager, err error) {
	mgr = &ManagerDocker{execCtx: &cm.ExecContext{}}

	return
}
