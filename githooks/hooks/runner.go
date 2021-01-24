package hooks

import (
	"os"
	"path"
	"regexp"
	cm "rycus86/githooks/common"
)

// The data for the runner config file.
type runnerConfigFile struct {
	Cmd  string   `yaml:"cmd"`
	Args []string `yaml:"args"`

	Version int `yaml:"version"`
}

// The current runner config gile version.
var runnerConfigFileVersion int = 1

// createHookIgnoreFile creates the data for the runner config file.
func createRunnerConfig() runnerConfigFile {
	return runnerConfigFile{Version: runnerConfigFileVersion}
}

// Load a runner configuration file.
func loadRunnerConfig(hookPath string) (data runnerConfigFile, err error) {
	data = createRunnerConfig()
	err = cm.LoadYAML(hookPath, &data)

	if data.Version == 0 {
		err = cm.ErrorF("Version '%v' needs to be greater than 0.", data.Version)

		return
	}

	return
}

// GetHookRunCmd gets the executable for the hook `hookPath`.
func GetHookRunCmd(hookPath string, args []string) (exec cm.Executable, err error) {

	if cm.IsExecutable(hookPath) {
		exec.Cmd = hookPath

		return
	}

	if path.Ext(hookPath) != ".yaml" {
		// No run configuration YAML -> get the default runner.
		exec = GetDefaultRunner(hookPath, args)

		return
	}

	config, e := loadRunnerConfig(hookPath)
	if e != nil {
		err = cm.ErrorF("Could not read runner config '%s'", hookPath)

		return
	}

	exec.Cmd = replaceEnvVariables(config.Cmd)
	exec.Args = config.Args

	for i := range config.Args {
		exec.Args[i] = replaceEnvVariables(exec.Args[i])
	}
	exec.Args = append(exec.Args, args...)

	return
}

var reEnvVariable = regexp.MustCompile(`\$?\$(\{[a-zA-Z]\w*\}|[a-zA-Z]\w*)`)

func replaceEnvVariables(s string) string {
	return reEnvVariable.ReplaceAllStringFunc(s, substituteEnvVariable)
}

func substituteEnvVariable(s string) string {
	r := []rune(s)

	if r[0] == '$' && r[1] == '$' {
		// Escape '$$var' or '$${var}' => '$var' or '${var}'
		return string(r[1:])
	}

	if r[1] == '{' {
		// Case: '${var}'
		return os.Getenv(string(r[2 : len(r)-1]))
	}

	// Case '$var'
	return os.Getenv(string(r[1:]))

}
