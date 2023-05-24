package images

// ContainerMgr provides the interface to `docker` or `podman`
// for the functionality used in Githooks.
// We do not use moby/moby because we would need to wrap agnostic arguments.
type IManager interface {
	ImagePull(image string) error
	ImageTag(imageSrc string, imageTarget string) error
	ImageBuild(dockerfile string, context string, target string) error
}
