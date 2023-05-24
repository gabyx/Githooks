package container

// ContainerMgr provides the interface to `docker` or `podman`
// for the functionality used in Githooks.
// We do not use moby/moby because we would need to wrap agnostic arguments.
type IManager interface {
	ImagePull(ref string) error
	ImageTag(refSrc string, refTarget string) error
	ImageBuild(dockerfile string, context string, target string, ref string) error
	ImageExists(ref string) (bool, error)
}
