package hooks

import (
	"io"
	"os"
	"testing"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/stretchr/testify/assert"
)

func TestLoadImagesConfig(t *testing.T) {
	file, err := os.CreateTemp("", "image.yaml")
	cm.AssertNoErrorPanic(err, "Could not create file.")
	defer os.Remove(file.Name())

	c := createImageConfigFile()
	cc := ImageConfig{}
	cc.Build = &ImageConfigBuild{Dockerfile: "thisfile", Target: "stage-1"}
	c.Images["bla"] = cc

	err = cm.StoreYAML(file.Name(), c)
	cm.AssertNoErrorPanic(err, "Could not store yaml.")
	config, err := loadImagesConfigFile(file.Name())
	cm.AssertNoErrorPanic(err, "Could not load yaml.")

	assert.Equal(t, config.Images["bla"].Build.Dockerfile, "thisfile")
	assert.Nil(t, config.Images["bla"].Pull)
}

func TestLoadImagesConfig2(t *testing.T) {
	file, err := os.CreateTemp("", "image.yaml")
	cm.AssertNoErrorPanic(err, "Could not create file.")
	defer os.Remove(file.Name())
	c := createImageConfigFile()

	cc := ImageConfig{}
	cc.Pull = &ImageConfigPull{Reference: "container:1.2@sha256:abf"}
	cc.Build = &ImageConfigBuild{Dockerfile: "thisfile", Target: "stage-1"}
	c.Images["bla"] = cc

	err = cm.StoreYAML(file.Name(), c)
	cm.AssertNoErrorPanic(err, "Could not store yaml.")
	config, err := loadImagesConfigFile(file.Name())
	cm.AssertNoErrorPanic(err, "Could not load yaml.")

	assert.Equal(t, config.Images["bla"].Build.Dockerfile, "thisfile")
	assert.Equal(t, config.Images["bla"].Pull.Reference, "container:1.2@sha256:abf")
}

func TestLoadImagesConfig3(t *testing.T) {
	file, err := os.CreateTemp("", "image.yaml")
	cm.AssertNoErrorPanic(err, "Could not create file.")

	content := `
version: 1
images:
`

	_, err = io.WriteString(file, content)
	cm.AssertNoErrorPanic(err, "Could not write file.")
	defer os.Remove(file.Name())

	config, err := loadImagesConfigFile(file.Name())
	cm.AssertNoErrorPanic(err, "Could not load yaml.")
	assert.Nil(t, config.Images)
}
