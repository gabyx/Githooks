package hooks

import (
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
	// cc.Pull = &ImageConfigPull{Name: "container", Tag: "1.2.9", Digest: "sha256:123"}
	cc.Build = &ImageConfigBuild{File: "thisfile", Target: "stage-1"}
	c.Images["bla"] = cc

	err = cm.StoreYAML(file.Name(), c)
	cm.AssertNoErrorPanic(err, "Could not store yaml.")
	config, err := loadImagesConfigFile(file.Name())
	cm.AssertNoErrorPanic(err, "Could not load yaml.")

	assert.Equal(t, config.Images["bla"].Build.File, "thisfile")
	assert.Nil(t, config.Images["bla"].Pull)

	c = createImageConfigFile()
	cc = ImageConfig{}
	cc.Pull = &ImageConfigPull{Name: "container", Tag: "1.2.9", Digest: "sha256:123"}
	cc.Build = &ImageConfigBuild{File: "thisfile", Target: "stage-1"}
	c.Images["bla"] = cc

	err = cm.StoreYAML(file.Name(), c)
	cm.AssertNoErrorPanic(err, "Could not store yaml.")
	config, err = loadImagesConfigFile(file.Name())
	cm.AssertNoErrorPanic(err, "Could not load yaml.")

	assert.Equal(t, config.Images["bla"].Build.File, "thisfile")
	assert.Equal(t, config.Images["bla"].Pull.Name, "container")
}
