package hooks

import (
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
)

type ImageConfigPull struct {
	Name   string `yaml:"name"`
	Tag    string `yaml:"tag"`
	Digest string `yaml:"digest"`
}

type ImageConfigBuild struct {
	File   string `yaml:"file"`
	Target string `yaml:"target"`
}

type ImageConfig struct {
	Pull  *ImageConfigPull  `yaml:"pull"`
	Build *ImageConfigBuild `yaml:"build"`
}

// ImagesConfigFile is the format of the images config file.
type ImagesConfigFile struct {
	// All configured images.
	Images map[string]ImageConfig `yaml:"images"`

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
func GetRepoImagesFile(repoDir string) string {
	return path.Join(GetGithooksDir(repoDir), ".images.yaml")
}
