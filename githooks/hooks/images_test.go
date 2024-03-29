package hooks

import (
	"io"
	"os"
	"path"
	"testing"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/container"
	"github.com/stretchr/testify/assert"
)

func TestLoadImagesConfig(t *testing.T) {
	file, err := os.CreateTemp("", "image.yaml")
	assert.Nil(t, err)
	defer os.Remove(file.Name())

	c := createImageConfigFile()
	cc := ImageConfig{}
	cc.Build = &ImageConfigBuild{Dockerfile: "thisfile", Stage: "stage-1"}
	c.Images["test-image"] = cc

	err = cm.StoreYAML(file.Name(), c)
	assert.Nil(t, err)
	config, err := loadImagesConfigFile(file.Name())
	assert.Nil(t, err)

	assert.Equal(t, config.Images["test-image"].Build.Dockerfile, "thisfile")
	assert.Nil(t, config.Images["test-image"].Pull)
}

func TestLoadImagesConfig2(t *testing.T) {
	file, err := os.CreateTemp("", "image.yaml")
	assert.Nil(t, err)
	defer os.Remove(file.Name())
	c := createImageConfigFile()

	cc := ImageConfig{}
	cc.Pull = &ImageConfigPull{Reference: "container:1.2@sha256:abf"}
	cc.Build = &ImageConfigBuild{Dockerfile: "thisfile", Stage: "stage-1"}
	c.Images["test-image"] = cc

	err = cm.StoreYAML(file.Name(), c)
	assert.Nil(t, err)
	config, err := loadImagesConfigFile(file.Name())
	assert.Nil(t, err)

	assert.Equal(t, config.Images["test-image"].Build.Dockerfile, "thisfile")
	assert.Equal(t, config.Images["test-image"].Pull.Reference, "container:1.2@sha256:abf")
}

func TestLoadImagesConfig3(t *testing.T) {
	file, err := os.CreateTemp("", "image.yaml")
	assert.Nil(t, err)

	content := `
version: 1
images:
`

	_, err = io.WriteString(file, content)
	assert.Nil(t, err)
	defer os.Remove(file.Name())

	config, err := loadImagesConfigFile(file.Name())
	assert.Nil(t, err)
	assert.Nil(t, config.Images)
}

func TestUpdateImages(t *testing.T) {

	repo, err := os.MkdirTemp("", "repo")
	assert.Nil(t, err)
	defer os.RemoveAll(repo)

	err = os.MkdirAll(path.Join(repo, ".githooks/docker/src"), cm.DefaultFileModeDirectory)
	assert.Nil(t, err)

	err = os.WriteFile(path.Join(repo, ".githooks/.namespace"), []byte("mynamespace"), cm.DefaultFileModeFile)
	assert.Nil(t, err)

	imageConfig := path.Join(repo, ".githooks/.images.yaml")

	content := []byte(`
version: 1
images:
  registry.com/${namespace}-test-image:mine1:
    pull:
      reference: alpine:latest

  test-image:mine2:
    pull:
      reference: alpine:3.16

  registry.com/dir/test-image:mine3:
    build:
      dockerfile: ./.githooks/docker/Dockerfile
      stage: stage2
      context: ./.githooks/docker/src
`)

	err = os.WriteFile(imageConfig, content, cm.DefaultFileModeFile)
	assert.Nil(t, err)

	err = os.WriteFile(path.Join(repo, ".githooks/docker/src/test"), nil, cm.DefaultFileModeFile)
	assert.Nil(t, err)

	content = []byte(`
FROM alpine:3.16 as stage1
COPY test /test-file

FROM stage1 as stage2
RUN apk add bash
`)

	err = os.WriteFile(path.Join(repo, ".githooks/docker/Dockerfile"), content, cm.DefaultFileModeFile)
	assert.Nil(t, err)

	log, err := cm.CreateLogContext(false, false)
	assert.Nil(t, err)

	mgr, err := container.NewManager("docker", nil)
	assert.NotNil(t, err)
	err = UpdateImages(log, "test-repo", repo, path.Join(repo, ".githooks"), "", mgr)
	assert.Nil(t, err, "Update images failed: %s", err)

	mgr, err = container.NewManager("", nil)
	assert.Nil(t, err)

	// Check all images.
	exists, err := mgr.ImageExists("registry.com/mynamespace-test-image:mine1")
	assert.Nil(t, err)
	assert.True(t, exists)

	exists, err = mgr.ImageExists("test-image:mine2")
	assert.Nil(t, err)
	assert.True(t, exists)

	exists, err = mgr.ImageExists("registry.com/dir/test-image:mine3")
	assert.Nil(t, err)
	assert.True(t, exists)

	// Remove all images.
	err = mgr.ImageRemove("registry.com/mynamespace-test-image:mine1")
	assert.Nil(t, err)
	err = mgr.ImageRemove("test-image:mine2")
	assert.Nil(t, err)
	err = mgr.ImageRemove("registry.com/dir/test-image:mine3")
	assert.Nil(t, err)

}
