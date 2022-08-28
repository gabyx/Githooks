package hooks

import (
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
)

type NamespaceEnvs map[string][]string

// Get returns all env. variables to a corresponding namespace.
func (e NamespaceEnvs) Get(namespace string) []string {
	if e != nil {
		return e[namespace]
	}

	return nil
}

// The `envs.yaml` config which defines per-namespace
// environment variables for executing hooks in certain namespaces.
type namespaceEnvsConfig struct {
	NamespaceEnvs NamespaceEnvs `yaml:"envs"`
	// The version of the file.
	Version int `yaml:"version"`
}

// Version for sharedHookConfig.
// Version 1: Initial.
const namespaceEnvsConfigVersion int = 1

func createNamespaceEnvsConfig() namespaceEnvsConfig {
	return namespaceEnvsConfig{Version: namespaceEnvsConfigVersion}
}

func GetEnvFile(repoHooksDir string) string {
	return path.Join(repoHooksDir, "envs.yaml")
}

// LoadNamespaceEnvs loads the envs config file in the repository if existing.
func LoadNamespaceEnvs(repoHooksDir string) (namespaceEnvs NamespaceEnvs, err error) {
	config := createNamespaceEnvsConfig()
	file := GetEnvFile(repoHooksDir)

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

	return config.NamespaceEnvs, err
}
