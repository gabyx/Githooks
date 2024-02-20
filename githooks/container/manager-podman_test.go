//go:build !test_docker && test_podman

package container

import (
	"testing"
)

func TestPodmanManager(t *testing.T) {
	testDockerManager(t, "podman")
}

func TestPodmanManagerBuild(t *testing.T) {
	testDockerManagerBuild(t, "podman")
}

func TestPodmanManagerBuildFail(t *testing.T) {
	testDockerManagerBuildFail(t, "podman")
}
