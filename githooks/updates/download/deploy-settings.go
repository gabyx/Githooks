package download

import (
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
)

// DeploySettings are the settings a user of Githooks can adjust to
// successfully download updates.
type deploySettings struct {
	Version int `yaml:"version"`

	Gitea  *GiteaDeploySettings  `yaml:"gitea"`
	Github *GithubDeploySettings `yaml:"github"`
	HTTP   *HTTPDeploySettings   `yaml:"http"`
	Local  *LocalDeploySettings  `yaml:"local"`
}

const deploySettingsVersion = 1

// IDeploySettings is the common interface for all deploy settings.
type IDeploySettings interface {
	Download(versionTag string, dir string) error
}

// LoadDeploySettings load the deploy settings from `file`.
func LoadDeploySettings(file string) (IDeploySettings, error) {
	var settings deploySettings
	if err := cm.LoadYAML(file, &settings); err != nil {
		return nil, err
	}

	if settings.Version == 0 {
		return nil, cm.ErrorF("Key 'Version' > 0 needs to be specified.")
	}

	switch {
	case settings.Gitea != nil:
		return settings.Gitea, nil
	case settings.Github != nil:
		return settings.Github, nil
	case settings.HTTP != nil:
		return settings.HTTP, nil
	case settings.Local != nil:
		return settings.Local, nil
	}

	return nil, nil
}

// StoreDeploySettings stores the deploy `settings` to `file`.
func StoreDeploySettings(file string, settings IDeploySettings) error {

	var s deploySettings

	// Always store the new version
	s.Version = deploySettingsVersion

	switch v := settings.(type) {
	case *GiteaDeploySettings:
		s.Gitea = v
	case *GithubDeploySettings:
		s.Github = v
	case *HTTPDeploySettings:
		s.HTTP = v
	case *LocalDeploySettings:
		s.Local = v
	default:
		cm.PanicF("Cannot store deploy settings for type '%T'", v)
	}

	return cm.StoreYAML(file, &s)
}

// GetDeploySettingsFile gets the deploy settings file inside the install directory.
func GetDeploySettingsFile(installDir string) string {
	return path.Join(installDir, "deploy.yaml")
}
