package container

import (
	"io"
	"os"
	"testing"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/stretchr/testify/assert"
)

func TestDockerManager(t *testing.T) {
	mgr, err := NewManager("docker")
	assert.Nil(t, err)

	err = mgr.ImagePull("alpine:latest")
	assert.Nil(t, err, "Could not pull image: %s", err)

	err = mgr.ImagePull("alpine:latests")
	assert.NotNil(t, err, "Pull image should have failed: %s", err)

	err = mgr.ImageTag("alpine:latest", "alpine:mine")
	assert.Nil(t, err, "Tagging image should not have failed: %s", err)

	exists, err := mgr.ImageExists("alpine:latest")
	assert.Nil(t, err)
	assert.True(t, exists)

	exists, err = mgr.ImageExists("alpine:mine")
	assert.Nil(t, err)
	assert.True(t, exists)

	exists, err = mgr.ImageExists("alpine:latests")
	assert.Nil(t, err)
	assert.False(t, exists)

	err = mgr.ImageRemove("alpine:latest")
	assert.Nil(t, err)
}

func TestDockerManagerBuild(t *testing.T) {
	mgr, err := NewManager("docker")
	assert.Nil(t, err)

	file, err := os.CreateTemp("", "")
	assert.Nil(t, err)
	defer os.Remove(file.Name())
	dockerfile := `
FROM alpine:latest as stage1

FROM stage1 as stage2
RUN apk add bash
`
	_, _ = io.WriteString(file, dockerfile)
	file.Close()

	log, err := cm.CreateLogContext(false)
	assert.Nil(t, err)

	exists, err := mgr.ImageExists("alpine:mine-special")
	assert.Nil(t, err)
	assert.False(t, exists)

	_, err = mgr.ImageBuild(log, file.Name(), ".", "stage2", "alpine:mine-special")
	assert.Nil(t, err, "Build failed: '%s'", err)

	exists, err = mgr.ImageExists("alpine:mine-special")
	assert.Nil(t, err)
	assert.True(t, exists)

	err = mgr.ImageRemove("alpine:mine-special")
	assert.Nil(t, err)
}

func TestDockerManagerBuildFail(t *testing.T) {
	mgr, err := NewManager("docker")
	assert.Nil(t, err)

	file, err := os.CreateTemp("", "")
	assert.Nil(t, err)
	defer os.Remove(file.Name())
	dockerfile := `
FROM alpine:latest as stage1

FROM stage1 as stage2
RUN apk add bashhhh
`
	_, _ = io.WriteString(file, dockerfile)
	file.Close()

	log, err := cm.CreateLogContext(false)
	assert.Nil(t, err)

	_, err = mgr.ImageBuild(log, file.Name(), ".", "stage2", "alpine:mine-special")
	assert.NotNil(t, err, "Build failed: '%s'", err)

	exists, err := mgr.ImageExists("alpine:mine-special")
	assert.Nil(t, err)
	assert.False(t, exists)
}
