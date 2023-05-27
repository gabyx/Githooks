package hooks

import (
	"io"
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/container"
	"github.com/gabyx/githooks/githooks/git"
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
	fromHint string,
	repositoryDir string,
	hooksDir string) (err error) {

	file := GetRepoImagesFile(hooksDir)

	if exists, _ := cm.IsPathExisting(file); !exists {
		log.Debug("No images config existing. Skip updating images.")

		return
	}

	log.InfoF("Building images for '%s'...", fromHint)

	gitx := git.NewCtx()
	manager := gitx.GetConfig(GitCKContainerManager, git.Traverse)
	mgr, e := container.NewManager(manager)
	if e != nil {
		err = cm.CombineErrors(cm.Error("Creating container manager failed."), e)

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
					"in '.images.yaml' in '%s' will be ignored\n"+
					"because pull is specified.", name, file)

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
			if path.IsAbs(img.Build.Context) {
				err = cm.Error(
					"Build context path '%s' given in '%s' must be a relative path.",
					img.Build.Context, file)

				return
			}

			if path.IsAbs(img.Build.Dockerfile) {
				err = cm.ErrorF(
					"Dockerfile path '%s' given in '%s' must be a relative path.",
					img.Build.Dockerfile, file)

				return
			}

			err = mgr.ImageBuild(
				log,
				path.Join(repositoryDir, img.Build.Dockerfile),
				path.Join(repositoryDir, img.Build.Context),
				img.Build.Target, name)

			if err != nil {
				// Save build error to temporary file.
				file, _ := os.CreateTemp("", "githooks-image-build-error-*.log")
				defer file.Close()
				_, e = io.WriteString(file, err.Error())
				log.AssertNoError(e, "Could not save image build errors.")

				err = cm.ErrorF("Building image '%s' from '%s' did not succeed.\n"+
					"Inspect build errors in file '%s'.",
					pullSrc, name, file.Name())

				return
			}
		}

	}

	return
}
