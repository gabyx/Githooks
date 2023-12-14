package hooks

import (
	"path/filepath"
	"runtime"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/container"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"

	"os"
	"path"

	"github.com/agext/regexp"
)

type imageRunConfig struct {
	Reference string `yaml:"reference"`
}

// The data for the runner config file.
type runnerConfigFile struct {
	Cmd   string         `yaml:"cmd"`
	Args  []string       `yaml:"args"`
	Env   []string       `yaml:"env"`
	Image imageRunConfig `yaml:"image"`

	Version int `yaml:"version"`
}

// The current runner config gile version.
// Version 1: Initial file.
// Version 2: Added `Env` field.
// Version 3: Added `Images` field.
var runnerConfigFileVersion int = 3

// createHookIgnoreFile creates the data for the runner config file.
func createRunnerConfig() runnerConfigFile {
	return runnerConfigFile{Version: runnerConfigFileVersion}
}

// Load a runner configuration file.
func loadRunnerConfig(hookPath string) (data runnerConfigFile, err error) {
	data = createRunnerConfig()
	err = cm.LoadYAML(hookPath, &data)

	if data.Version < 0 || data.Version > runnerConfigFileVersion {
		err = cm.ErrorF(
			"File '%s' has version '%v'. "+
				"This version of Githooks only supports version >= 1 and <= '%v'.",
			hookPath,
			data.Version,
			runnerConfigFileVersion)

		return
	}

	return
}

// GetHookRunCmd gets the executable for the hook `hookPath`.
// Any command in a runner config YAML with path separators will
// be made absolute to `rootDir`.
func GetHookRunCmd(
	gitx *git.Context,
	hookPath string,
	rootDir string,
	hooksDir string,
	parseRunnerConfig bool,
	containerizedEnabled bool,
	hookNamespace string,
	envs []string) (cm.IExecutable, error) {

	exec := cm.NewExecutable(hookPath, nil, envs)

	if cm.IsExecutable(exec.Cmd) {
		return &exec, nil
	}

	if !parseRunnerConfig || path.Ext(hookPath) != ".yaml" {
		// Dont parse run config or not existing -> get the default runner.
		return GetDefaultRunner(hookPath, envs), nil
	}

	config, e := loadRunnerConfig(hookPath)
	if e != nil {
		return nil, cm.ErrorF("Could not read runner config '%s'", hookPath)
	}

	subst := getVarSubstitution(os.LookupEnv, gitx.LookupConfig)

	// Substitute variable in env values.
	var err error
	for i := range config.Env {
		if config.Env[i], err = subst(config.Env[i]); err != nil {
			return nil, cm.CombineErrors(err,
				cm.ErrorF("Error in hook run config '%s'.", hookPath))
		}
	}

	// Substitute variables in command.
	if exec.Cmd, err = subst(config.Cmd); err != nil {
		return nil, cm.CombineErrors(err,
			cm.ErrorF("Error in hook run config '%s'.", hookPath))
	}

	exec.Args = config.Args

	// Substitute variables in arguments.
	for i := range exec.Args {
		if exec.Args[i], err = subst(exec.Args[i]); err != nil {
			return nil, cm.CombineErrors(err,
				cm.ErrorF("Error in hook run config '%s'.", hookPath))
		}
	}

	exec.Env = append(exec.Env, config.Env...)
	exec.Env = append(exec.Env, envs...)

	if containerizedEnabled && strs.IsNotEmpty(config.Image.Reference) {
		// Containerized execution.

		manager := gitx.GetConfig(GitCKContainerManager, git.Traverse)
		mgr, err := container.NewManager(manager)
		if err != nil {
			return nil, cm.CombineErrors(err, cm.Error("Could not create container manager."))
		}

		reference, err := addImageReferenceSuffix(config.Image.Reference, hookPath, hookNamespace)
		if err != nil {
			return nil, err
		}

		containerExec, err := mgr.NewHookRunExec(reference, gitx.GetCwd(), rootDir, &exec)

		if err != nil {
			return nil, cm.CombineErrors(err, cm.Error("Could not create container hook executor."))
		}

		return containerExec, nil
	} else {
		// Normal execution.

		// Resolve commands with path separators which are
		// relative paths relative to the `rootDir`.
		// e.g `dist/custom.exe` -> `rootDir/dist/custom.exe`
		if strings.ContainsAny(exec.Cmd, "/\\") {
			if runtime.GOOS == cm.WindowsOsName {
				exec.Cmd = filepath.ToSlash(exec.Cmd)
			}

			if !filepath.IsAbs(exec.Cmd) {
				exec.Cmd = path.Join(rootDir, exec.Cmd)
			}
		}

		return &exec, nil
	}
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
				subs, exists = gitGet(match[4], git.SystemScope)
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
