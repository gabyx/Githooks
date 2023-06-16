package hooks

import (
	"io"
	"os"
	"path"
	"strings"

	ref "github.com/distribution/distribution/reference"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/container"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
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
	// The optional stage in the dockerfile which should be build.
	Stage string `yaml:"stage"`
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
	namespace, e := GetHooksNamespace(hooksDir)
	log.AssertNoError(e, "Could not get hooks namespace in '%s'.", hooksDir)

	nBuilds := 0
	nPulls := 0

	if exists, _ := cm.IsPathExisting(file); !exists {
		log.DebugF("No images config existing '%s'. Skip updating images.", file)

		return
	}

	log.InfoF("Build/pull images for repository '%s'...", fromHint)

	gitx := git.NewCtx()
	manager := gitx.GetConfig(GitCKContainerManager, git.Traverse)
	mgr, err := container.NewManager(manager)
	if err != nil {
		return cm.CombineErrors(cm.Error("Creating container manager failed."), err)
	}

	var imagesConfig ImagesConfigFile

	imagesConfig, err = loadImagesConfigFile(file)
	if err != nil {
		return cm.CombineErrors(
			cm.ErrorF("Could not load images config file '%s'.", file), err)
	}

	for imageRef, img := range imagesConfig.Images {

		imageRef, e := addImageReferenceSuffix(imageRef, file, namespace)
		if e != nil {
			err = cm.CombineErrors(err, e)

			continue
		}

		pullSrc := imageRef

		if img.Pull != nil {
			log.WarnIfF(img.Build != nil,
				"Specified image build configuration on entry '%s'\n"+
					"in '.images.yaml' in '%s' will be ignored\n"+
					"because pull is specified.", imageRef, file)

			pullSrc = img.Pull.Reference
		}

		if img.Build == nil {
			// Do a pull of the image, because `build` is not specified.

			e := mgr.ImagePull(pullSrc)

			if e != nil {
				err = cm.CombineErrors(err,
					cm.ErrorF("Pulling image '%s' in '%s' did not succeed.\n"+
						"Hooks may not run correctly.", imageRef, file), e)

				continue
			}

			log.InfoF(
				"  %s Pulled image '%s'.", cm.ListItemLiteral, pullSrc, fromHint)
			nPulls += 1

			if imageRef != pullSrc {
				e = mgr.ImageTag(pullSrc, imageRef)

				if e != nil {
					err = cm.CombineErrors(err,
						cm.ErrorF("Retagging image '%s' to '%s' in '%s' did not succeed.\n"+
							"Hooks may not run correctly.", pullSrc, imageRef, file), e)

					continue
				}

				log.InfoF(
					"  %s Tagged image '%s' to\n"+
						"     -> '%s'.", cm.ListItemLiteral, pullSrc, imageRef, fromHint)
			}
		} else if img.Pull == nil {
			// Do a build of the image because no `pull` but `build` specified.

			if path.IsAbs(img.Build.Context) {
				err = cm.CombineErrors(err, cm.Error(
					"Build context path '%s' given in '%s' must be a relative path.",
					img.Build.Context, file))

				continue
			}

			if path.IsAbs(img.Build.Dockerfile) {
				err = cm.CombineErrors(err, cm.ErrorF(
					"Dockerfile path '%s' given in '%s' must be a relative path.",
					img.Build.Dockerfile, file))

				continue
			}

			out, e := mgr.ImageBuild(
				log,
				path.Join(repositoryDir, img.Build.Dockerfile),
				path.Join(repositoryDir, img.Build.Context),
				img.Build.Stage, imageRef)

			if e != nil {
				// Save build error to temporary file.
				file, _ := os.CreateTemp("", "githooks-image-build-error-*.log")
				defer file.Close()
				_, e2 := io.WriteString(file,
					e.Error()+
						"\nOutput:\n=====================================================\n"+
						out)
				log.AssertNoError(e2, "Could not save image build errors.")

				const maxChars int = 500
				err = cm.CombineErrors(err,
					cm.ErrorF("Building image '%s' did not succeed.\n"+
						"Inspect build errors in file '%s' with\n"+
						"`cat '%s'\n"+
						"Partial output:\n%s",
						imageRef, file.Name(), file.Name(), out[0:cm.Min(len(out), maxChars)]))

				continue
			}

			log.InfoF("  %v Built image '%s'.", cm.ListItemLiteral, imageRef)
			nBuilds += 1
		}
	}

	log.InfoF("Pulled '%v' and built '%v' images.", nPulls, nBuilds)

	return
}

// addImageReferenceSuffix adds the `namespace` to a image name reference at the place `${namespace}`.
func addImageReferenceSuffix(imageRef string, file string, namespace string) (string, error) {
	if !strs.IsEmpty(namespace) {
		imageRef = strings.Replace(imageRef, "${namespace}", namespace, 1)
	}

	_, e := ref.Parse(imageRef)

	if e != nil {
		return imageRef,
			cm.ErrorF("Could not parse image reference."+
				"Image reference '%s' in '%s' must be a "+
				"named reference according to "+
				"'https://github.com/distribution/distribution/blob/main/reference/reference.go'", imageRef, file)
	}

	return imageRef, nil
}
