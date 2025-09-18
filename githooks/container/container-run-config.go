package container

import (
	"os"

	cm "github.com/gabyx/githooks/githooks/common"
)

// containerRunConfig is the format of the container run arguments file.
type containerRunConfig struct {
	// The path in the container where the workspace dir is located.
	// Defaults to `/mnt/workspace`.
	WorkspacePathDest string `yaml:"workspace-path-dest"`

	// If the workspace directory is automatically mounted or
	// you do it yourself with `Args`. Defaults to `true`.
	// Giving you the chance to mount it differently,
	// e.g. `--volumes-from other-container` when you do not
	// know the host path because you are already inside a container.
	AutoMountWorkspace bool `yaml:"auto-mount-workspace"`

	// The path in the container to the directory where all shared repositories
	// are located.
	// Defaults to `/mnt/shared`.
	SharedPathDest string `yaml:"shared-path-dest"`

	// If the shared directory is automatically mounted or
	// you do it yourself `Args`.
	// Giving you the chance to mount it differently,
	// e.g. `--volumes-from other-container` when you do not
	// know the host path because you are already inside a container.
	// Defaults to `true`.
	AutoMountShared bool `yaml:"auto-mount-shared"`

	// Additional arguments to the container run command.
	Args []string `yaml:"args"`

	// The version of this file format.
	Version int `yaml:"version"`
}

// Version for containerRunConfig.
// Version 1: Initial.
const containerRunConfigVersion int = 1

func createContainerRunConfig() containerRunConfig {
	return containerRunConfig{
		WorkspacePathDest:  "/mnt/workspace",
		AutoMountWorkspace: true,

		SharedPathDest:  "/mnt/shared",
		AutoMountShared: true,

		Version: containerRunConfigVersion,
	}
}

func loadContainerRunConfig() (config containerRunConfig, err error) {
	config = createContainerRunConfig()
	file, exists := os.LookupEnv(EnvVariableContainerRunConfigFile)

	if exists && cm.IsFile(file) {
		err = cm.LoadYAML(file, &config)
		if err != nil {
			err = cm.CombineErrors(err, cm.ErrorF("Could not load file '%s'", file))

			return
		}

		if config.Version < 0 || config.Version > containerRunConfigVersion {
			err = cm.ErrorF(
				"File '%s' has version '%v'. "+
					"This version of Githooks only supports version >= 1 and <= '%v'.",
				file,
				config.Version,
				containerRunConfigVersion)

			return
		}
	}

	return config, nil
}
