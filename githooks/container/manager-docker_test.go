//go:build test_docker && !test_podman

package container

import (
	"testing"
)

func TestDockerManager(t *testing.T) {
	testDockerManager(t, "docker")
}

func TestDockerManagerBuild(t *testing.T) {
	testDockerManagerBuild(t, "docker")
}

func TestDockerManagerBuildFail(t *testing.T) {
	testDockerManagerBuildFail(t, "docker")
}
