package hooks

import (
	cm "gabyx/githooks/common"
	"gabyx/githooks/git"

	"os"
	"path"

	"github.com/agext/regexp"
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

	subst := getVarSubstitution(os.LookupEnv, git.Ctx().LookupConfig)

	if exec.Cmd, err = subst(config.Cmd); err != nil {
		err = cm.CombineErrors(err,
			cm.ErrorF("Error in hook run config '%s'.", hookPath))

		return
	}

	exec.Args = config.Args

	for i := range config.Args {
		if exec.Args[i], err = subst(exec.Args[i]); err != nil {
			err = cm.CombineErrors(err,
				cm.ErrorF("Error in hook run config '%s'.", hookPath))

			return
		}

	}
	exec.Args = append(exec.Args, args...)

	return
}

var reEnvVariable = regexp.MustCompile(`(\\?)\$\{(!?)(env|git|git-l|git-g|git-s):([a-zA-Z.][a-zA-Z0-9_.]+)\}`)

func getVarSubstitution(
	getEnv func(string) (string, bool),
	gitGet func(string, git.ConfigScope) (string, bool)) func(string) (string, error) {

	return func(s string) (res string, err error) {

		res = reEnvVariable.ReplaceAllStringSubmatchFunc(s, func(match []string) (subs string) {

			// Escape '\${var}' => '${var}'
			if len(match[1]) != 0 {
				return string([]rune(match[0])[1:])
			}

			var exists bool

			switch match[3] {
			case "env":
				subs, exists = getEnv(match[4])
			case "git":
				subs, exists = gitGet(match[4], git.Traverse)
			case "git-l":
				subs, exists = gitGet(match[4], git.LocalScope)
			case "git-g":
				subs, exists = gitGet(match[4], git.GlobalScope)
			case "git-s":
				subs, exists = gitGet(match[4], git.System)
			default:
				cm.DebugAssert(false, "This should not happen.")
			}

			if len(match[2]) != 0 && !exists {
				err = cm.ErrorF("Config variable '%s' could not be substituted\n"+
					"because it does not exist!", match[0])
			}

			return
		})

		return
	}
}
