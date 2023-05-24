package hooks

import (
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/container"
)

type ImageConfigPull struct {
	// The image reference to pull.
	// See https://github.com/distribution/distribution/blob/main/reference/reference.go
	Reference string `yaml:"reference"`
}

type ImageConfigBuild struct {
	// The docker file for the build.
	Dockerfile string `yaml:"dockerfile"`
	// The optional context directory relative to the shared Githooks repository.
	Context string `yaml:"context"`
	// The optional target in the dockerfile which should be build.
	Target string `yaml:"target"`
}

type ImageConfig struct {
	Pull  *ImageConfigPull  `yaml:"pull"`
	Build *ImageConfigBuild `yaml:"build"`
}

// ImagesConfig represents pull/build config settings for
// a specified image reference.
type ImagesConfig = map[string]ImageConfig

// ImagesConfigFile is the format of the images config file.
type ImagesConfigFile struct {
	// All configured images.
	Images ImagesConfig `yaml:"images"`

	// The version of the file.
	Version int `yaml:"version"`
}

// Version for ImagesConfigFile.
// Version 1: Initial.
const imagesConfigFileVersion int = 1

func createImageConfigFile() ImagesConfigFile {
	return ImagesConfigFile{Version: imagesConfigFileVersion, Images: make(map[string]ImageConfig)}
}

func loadImagesConfigFile(file string) (config ImagesConfigFile, err error) {
	config = ImagesConfigFile{}

	if cm.IsFile(file) {
		err = cm.LoadYAML(file, &config)
		if err != nil {
			err = cm.CombineErrors(err, cm.ErrorF("Could not load file '%s'", file))

			return
		}

		if config.Version == 0 {
			err = cm.ErrorF("Version '%v' needs to be greater than 0.", config.Version)

			return
		}
	}

	return config, nil
}

// GetRepoSharedFile gets the shared file with respect to the hooks dir in the repository.
func GetRepoImagesFile(hookDir string) string {
	return path.Join(hookDir, ".images.yaml")
}

// UpdateImages updates the images from the `images` config from the
// `hooksDir` inside `repositoryDir` (can be shared) by pulling or building them.
func UpdateImages(
	log cm.ILogContext,
	repositoryDir string,
	hooksDir string,
	mgr container.IManager) (err error) {

	file := GetRepoImagesFile(hooksDir)

	if exists, _ := cm.IsPathExisting(file); !exists {
		log.Debug("No images config existing. Skip updating images.")

		return
	}

	var imagesConfig ImagesConfigFile

	imagesConfig, err = loadImagesConfigFile(file)
	log.AssertNoErrorPanic(err,
		"Could not load images config file '%s'.", file)

	for name, img := range imagesConfig.Images {
		pullSrc := name

		if img.Pull != nil {
			log.WarnIfF(img.Build != nil,
				"Specified image build configuration on entry '%s'\n"+
					"will be ignored because pull is specified.", name)

			pullSrc = img.Pull.Reference
		}

		err = mgr.ImagePull(pullSrc)

		if err != nil {
			err = cm.CombineErrors(err,
				cm.ErrorF("Pulling image '%s' did not succeed.\n"+
					"Hooks may not run correctly:\n%s",
					pullSrc, err))

			return
		}

		if name != pullSrc {
			err = mgr.ImageTag(pullSrc, name)

			if err != nil {
				err = cm.CombineErrors(err,
					cm.ErrorF("Retagging image '%s' to '%s' did not succeed.\n"+
						"Hooks may not run correctly:\n%s",
						pullSrc, name, err))

				return
			}

		}

		if img.Pull == nil && img.Build != nil {

			err = mgr.ImageBuild(img.Build.Dockerfile, img.Build.Context, img.Build.Target, name)

			if err != nil {
				err = cm.CombineErrors(err,
					cm.ErrorF("Building image '%s' from '%s' did not succeed.\n"+
						"Hooks may not run correctly:\n%s",
						pullSrc, name, err))

				return
			}
		}

	}

	return
}
