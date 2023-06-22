package container

import (
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gabyx/githooks/githooks/build"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

const (
	dockerCmd = "docker"
)

type ManagerDocker struct {
	cmdCtx cm.CmdContext

	uid string
	gid string
}

// ImagePull pulls an image with reference `ref`.
func (m *ManagerDocker) ImagePull(ref string) (err error) {
	return m.cmdCtx.Check("pull", ref)
}

// ImageTag tags an image with reference `refSrc` to reference `refTarget`.
func (m *ManagerDocker) ImageTag(refSrc string, refTarget string) (err error) {
	return m.cmdCtx.Check("tag", refSrc, refTarget)
}

// ImageBuild builds the stage `stage`
// of an image from `dockerfile` in context path `context` and tags
// it with reference `ref`.
func (m *ManagerDocker) ImageBuild(
	log cm.ILogContext,
	dockerfile string,
	context string,
	stage string,
	ref string) (string, error) {

	return m.cmdCtx.GetCombined(
		"build",
		"-f", dockerfile,
		"-t", ref,
		"--label", strs.Fmt("githooks-version=%v", build.GetBuildVersion().String()),
		"--target",
		stage, context)
}

// ImageExists checks if the image with reference `ref` exists.
func (m *ManagerDocker) ImageExists(ref string) (exists bool, err error) {
	out, err := m.cmdCtx.GetSplit("image", "ls", "--format", "{{ .ID }}", ref)

	return len(out) != 0, err
}

// ImageRemove removes an image with reference `ref`.
func (m *ManagerDocker) ImageRemove(ref string) (err error) {
	return m.cmdCtx.Check("image", "rm", ref)
}

func resolveWSBasePath(envValue string, dirname string) string {
	return strings.ReplaceAll(envValue, "${repository-dir-name}", dirname)
}

// NewHookRunExec runs a hook over a container.
func (m *ManagerDocker) NewHookRunExec(
	ref string,
	workspaceDir string,
	workspaceHookDir string,
	hookExec cm.IExecutable,
) (cm.IExecutable, error) {
	containerExec := ContainerizedExecutable{containerType: ContainerManagerTypeV.Docker}

	containerExec.Cmd = dockerCmd

	// Mount: Working directory.
	// The repository where the hook runs.
	mntWSSrc := workspaceDir
	mntWSDest := "/mnt/workspace"
	mntWSSharedSrc := workspaceHookDir
	mntWSSharedDest := "/mnt/shared"

	if hostPath := os.Getenv(EnvVariableContainerWorkspaceHostPath); strs.IsNotEmpty(hostPath) {
		containerExec.usedVolumes = true
		mntWSSrc = hostPath
	}

	workingDir := path.Join(mntWSDest,
		resolveWSBasePath(
			os.Getenv(EnvVariableContainerWorkspaceBasePath),
			path.Base(workspaceDir),
		)) // defaults to `mntWSDest`

	// Mount: Shared hook repository:
	// This mount contains a shared repository root directory.
	var cmdBasePath string
	mountWSShared := workspaceDir != workspaceHookDir

	if !mountWSShared {
		// Hooks are configured in current repository: Dont mount the shared location.
		cmdBasePath = mntWSDest
	} else {
		// Mount shared too.
		if hostPath := os.Getenv(EnvVariableContainerSharedHostPath); strs.IsNotEmpty(hostPath) {
			mntWSSharedSrc = hostPath
		} else if containerExec.usedVolumes {
			return nil, cm.ErrorF(
				"Host path for workspace '%s' set but missing a host path "+
					"for shared hooks to run containerized. "+
					"See the Githooks manual to configure it.", mntWSSrc)
		}

		cmdBasePath = path.Join(mntWSSharedDest, path.Base(workspaceHookDir))
	}

	// Resolve commands with path separators which are
	// relative paths relative to `cmdBasePath`.
	// e.g `dist/custom.exe` -> `rootDir/dist/custom.exe`
	cmd := hookExec.GetCommand()
	if strings.ContainsAny(hookExec.GetCommand(), "/\\") {
		if runtime.GOOS == cm.WindowsOsName {
			cmd = filepath.ToSlash(cmd)
		}

		if filepath.IsAbs(cmd) {
			return nil, cm.ErrorF("Command '%s' specified in '%s' must only contain relative paths "+
				"for running containerized.", cmd, workspaceHookDir)
		}
		cmd = path.Join(cmdBasePath, cmd)
	}

	cm.DebugAssertF(!strings.Contains(workspaceDir, "\\"),
		"No forward slashes should be passed in here '%s'.", workspaceDir)
	cm.DebugAssertF(!strings.Contains(workspaceHookDir, "\\"),
		"No forward slashes should be passed in here '%s'.", workspaceHookDir)

	containerExec.ArgsPre = []string{
		"run",
		"--rm",
		"-v",
		strs.Fmt("%v:%v", mntWSSrc, mntWSDest), // Set the mount for the working directory.
		"-w", workingDir,                       // Set working dir.
	}

	if mountWSShared {
		containerExec.ArgsPre = append(containerExec.ArgsPre,
			"-v",
			strs.Fmt("%v:%v:ro", mntWSSharedSrc, mntWSSharedDest)) // Set the mount for the shared directory.
	}

	if runtime.GOOS != cm.WindowsOsName &&
		runtime.GOOS != "darwin" {
		// On non win/mac, execute as the user/group from the host.
		containerExec.ArgsPre = append(containerExec.ArgsPre,
			"--user",
			strs.Fmt("%v:%v", m.uid, m.gid))
	}

	// Set env. variable denoting we are running over a container.
	containerExec.ArgsEnv = []string{
		"-e", strs.Fmt("%s=true", EnvVariableContainerRun),
	}

	// Re-export env variables (does not contain general environment).
	for _, envKeyVar := range hookExec.GetEnvironment() {
		containerExec.ArgsEnv = append(containerExec.ArgsEnv, "-e", envKeyVar)
	}

	containerExec.ArgsPost = append(containerExec.ArgsPost, ref, cmd)
	containerExec.ArgsPost = append(containerExec.ArgsPost, hookExec.GetArgs()...)

	return &containerExec, nil
}

// IsDockerAvailable returns if docker is available.
func IsDockerAvailable() bool {
	_, err := exec.LookPath(dockerCmd)

	return err == nil
}

func NewManagerDocker() (mgr IManager, err error) {
	if !IsDockerAvailable() {
		return nil, &ManagerNotAvailableError{dockerCmd}
	}

	var uid, gid string

	if runtime.GOOS != cm.WindowsOsName && runtime.GOOS != "darwin" {
		usr, e := user.Current()

		if e != nil {
			err = cm.CombineErrors(err,
				cm.Error("Could not get user information for container manager."))

			return
		}

		uid = usr.Uid
		gid = usr.Gid
	}

	cmdCtx := cm.NewCommandCtxBuilder().SetBaseCmd(dockerCmd).EnableCaptureError().Build()
	mgr = &ManagerDocker{cmdCtx: cmdCtx, uid: uid, gid: gid}

	return
}
