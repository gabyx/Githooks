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

func pullImage(
	log cm.ILogContext,
	mgr container.IManager,
	pullSrc string,
	imageRef string,
	file string,
) (err error) {
	// Do a pull of the image, because `build` is not specified.
	err = mgr.ImagePull(pullSrc)

	if err != nil {
		err = cm.CombineErrors(
			cm.ErrorF("Pulling image '%s' in '%s' did not succeed.\n"+
				"Hooks may not run correctly.", imageRef, file), err)

		return
	}

	log.InfoF(
		"  %s Pulled image '%s'.", cm.ListItemLiteral, pullSrc)

	if imageRef != pullSrc {
		err = mgr.ImageTag(pullSrc, imageRef)

		if err != nil {
			err = cm.CombineErrors(
				cm.ErrorF("Retagging image '%s' to '%s' in '%s' did not succeed.\n"+
					"Hooks may not run correctly.", pullSrc, imageRef, file), err)

			return
		}

		log.InfoF(
			"  %s Tagged image '%s' to\n"+
				"     -> '%s'.", cm.ListItemLiteral, pullSrc, imageRef)
	}

	return
}

func buildImage(
	log cm.ILogContext,
	mgr container.IManager,
	context string,
	dockerfile string,
	stage string,
	imageRef string,
	file string,
	repositoryDir string) (err error) {
	// Do a build of the image because no `pull` but `build` specified.

	if path.IsAbs(context) {
		return cm.Error(
			"Build context path '%s' given in '%s' must be a relative path.",
			context, file)
	}

	if path.IsAbs(dockerfile) {
		return cm.ErrorF(
			"Dockerfile path '%s' given in '%s' must be a relative path.",
			dockerfile, file)
	}

	out, err := mgr.ImageBuild(
		log,
		path.Join(repositoryDir, dockerfile),
		path.Join(repositoryDir, context),
		stage,
		imageRef)

	if err != nil {
		// Save build error to temporary file.
		file, _ := os.CreateTemp("", "githooks-image-build-error-*.log")
		defer file.Close()
		_, e := io.WriteString(file,
			err.Error()+
				"\nOutput:\n=====================================================\n"+
				out)
		log.AssertNoError(e, "Could not save image build errors.")

		const maxChars int = 500
		length := cm.Min(len(out), maxChars)

		return cm.CombineErrors(err,
			cm.ErrorF("Building image '%s' did not succeed.\n"+
				"Inspect build errors in file '%s' with\n"+
				"`cat '%s'\n"+
				"Partial output:\n...stripped...\n%s",
				imageRef, file.Name(), file.Name(), out[len(out)-length:]))

	}

	log.InfoF("  %v Built image '%s'.", cm.ListItemLiteral, imageRef)

	return nil
}

// UpdateImages updates the images from the `images` config from the
// `hooksDir` inside `repositoryDir` (can be shared) by pulling or building them.
func UpdateImages(
	log cm.ILogContext,
	fromHint string,
	repositoryDir string,
	hooksDir string,
	configFile string) (err error) {

	if strs.IsEmpty(configFile) {
		configFile = GetRepoImagesFile(hooksDir)
	}

	namespace, e := GetHooksNamespace(hooksDir)
	log.AssertNoError(e, "Could not get hooks namespace in '%s'.", hooksDir)

	nBuilds := 0
	nPulls := 0

	if exists, _ := cm.IsPathExisting(configFile); !exists {
		log.InfoF("No images config existing '%s'. Skip updating images.", configFile)

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

	imagesConfig, err = loadImagesConfigFile(configFile)
	if err != nil {
		return cm.CombineErrors(
			cm.ErrorF("Could not load images config file '%s'.", configFile), err)
	}

	for imageRef, img := range imagesConfig.Images {

		imageRef, e := addImageReferenceSuffix(imageRef, configFile, namespace)
		if e != nil {
			err = cm.CombineErrors(err, e)

			continue
		}

		pullSrc := imageRef

		if img.Pull != nil {
			log.WarnIfF(img.Build != nil,
				"Specified image build configuration on entry '%s'\n"+
					"in '.images.yaml' in '%s' will be ignored\n"+
					"because pull is specified.", imageRef, configFile)

			pullSrc, e = addImageReferenceSuffix(img.Pull.Reference, configFile, namespace)

			if e != nil {
				err = cm.CombineErrors(err, e)

				continue
			}
		}

		if img.Build == nil {
			e := pullImage(
				log,
				mgr,
				pullSrc,
				imageRef,
				configFile)

			if e != nil {
				err = cm.CombineErrors(err, e)

				continue
			} else {
				nPulls += 1
			}

		} else if img.Pull == nil {
			e := buildImage(
				log,
				mgr,
				img.Build.Context,
				img.Build.Dockerfile,
				img.Build.Stage,
				imageRef,
				configFile,
				repositoryDir)

			if e != nil {
				err = cm.CombineErrors(err, e)

				continue
			} else {
				nBuilds += 1
			}

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
