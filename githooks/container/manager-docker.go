package container

import (
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

type ReadBindMount struct {
	Src  string
	Dest string
}

type ManagerDocker struct {
	cmdCtx cm.CmdContext

	uid string
	gid string

	// Only used to wrap podman into this structure as well.
	mgrType ContainerManagerType

	runConfig containerRunConfig
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
	cmd := []string{
		"build",
		"-f", dockerfile,
		"-t", ref,
		"--label", strs.Fmt("githooks-version=%v",
			build.GetBuildVersion().String()), //nolint:typecheck // Might not be generated yet.
	}

	if strs.IsNotEmpty(stage) {
		cmd = append(cmd, "--target", stage)
	}

	cmd = append(cmd, context)

	return m.cmdCtx.GetCombined(cmd...)
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

// NewHookRunExec runs a hook over a container.
func (m *ManagerDocker) NewHookRunExec(
	ref string,
	workspaceDir string,
	workspaceHookDir string,
	hookExec cm.IExecutable,
	attachStdIn bool,
	allocateTTY bool,
) (cm.IExecutable, error) {
	cm.DebugAssert(filepath.IsAbs(workspaceDir), "Workspace dir must be an absolute path.")
	cm.DebugAssert(filepath.IsAbs(workspaceHookDir), "Workspace hook dir must be an abs path.")

	containerExec := ContainerizedExecutable{containerType: m.mgrType}

	containerExec.Cmd = m.cmdCtx.GetBaseCmd()

	// Mount: Working directory.
	// The repository where the hook runs.
	// ==================================
	mntWSSrc := workspaceDir
	mntWSDest := m.runConfig.WorkspacePathDest
	// =================================

	// Mount: Shared hook directory.
	// The directory which contain all shared hooks repositories.
	// =================================
	mntWSSharedSrc := path.Dir(workspaceHookDir)
	mntWSSharedDest := m.runConfig.SharedPathDest
	// =================================

	workingDir := mntWSDest

	// Mount: Shared hook repository:
	// This mount contains a shared repository root directory.
	var cmdBasePath string
	// only mount shared if we need them (we are running shared hooks)
	mountWSShared := workspaceDir != workspaceHookDir

	if !mountWSShared {
		// Hooks are configured in current repository: Dont mount the shared location.
		cmdBasePath = mntWSDest
	} else {
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
			return nil, cm.ErrorF(
				"Command '%s' specified in '%s' must only contain relative paths "+
					"for running containerized.",
				cmd,
				workspaceHookDir,
			)
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
		"-w", workingDir, // Set working dir.
	}

	if attachStdIn {
		containerExec.ArgsPre = append(containerExec.ArgsPre, "--interactive")
	}

	if allocateTTY {
		containerExec.ArgsPre = append(containerExec.ArgsPre, "--tty")
	}

	// Mount the workspace directory if set.
	if m.runConfig.AutoMountWorkspace {
		containerExec.ArgsPre = append(containerExec.ArgsPre,
			"-v",
			strs.Fmt("%v:%v", mntWSSrc, mntWSDest), // Set the mount for the working directory.
		)
	}

	// Mount the shared directory if set.
	if m.runConfig.AutoMountShared && mountWSShared {
		containerExec.ArgsPre = append(
			containerExec.ArgsPre,
			"-v",
			strs.Fmt(
				"%v:%v:ro",
				mntWSSharedSrc,
				mntWSSharedDest,
			),
		) // Set the mount for the shared directory.
	}

	// Add all additional arguments.
	containerExec.ArgsPre = append(containerExec.ArgsPre, m.runConfig.Args...)

	switch m.mgrType {
	case ContainerManagerTypeV.Docker:
		if runtime.GOOS != cm.WindowsOsName &&
			runtime.GOOS != "darwin" {
			// On non win/mac, execute as the user/group from the host.
			// This will make all volume mounts have the same user/group
			// in the container.
			// The entrypoint https://github.com/FooBarWidget/matchhostfsowner
			// will take care to adjust a specified container user
			// to the one running.
			containerExec.ArgsPre = append(containerExec.ArgsPre,
				"--user",
				strs.Fmt("%v:%v", m.uid, m.gid))
		}
	case ContainerManagerTypeV.Podman:
		// With rootless podman its much easier to make the volumes
		// match the host user which launch this Githook.
		containerExec.ArgsPre = append(containerExec.ArgsPre, "--userns=keep-id:uid=1000,gid=1000")
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

func newManagerDocker(
	cmd string,
	mgrType ContainerManagerType,
	readMounts []ReadBindMount) (mgr *ManagerDocker, err error) {
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

	cmdCtx := cm.NewCommandCtxBuilder().SetBaseCmd(cmd).EnableCaptureError().Build()
	mgr = &ManagerDocker{cmdCtx: cmdCtx, uid: uid, gid: gid, mgrType: mgrType}

	// Load the run config or default it.
	mgr.runConfig, err = loadContainerRunConfig()
	if err != nil {
		err = cm.CombineErrors(
			err,
			cm.Error("Run config for containerized runs could not be loaded."),
		)
	}

	// Add additional mounts.
	for i := range readMounts {
		m := &readMounts[i]
		mgr.runConfig.Args = append(
			mgr.runConfig.Args,
			[]string{"-v", strs.Fmt("%s:%s:ro", m.Src, m.Dest)}...)
	}

	return
}

// NewManagerDocker return a new mangers for Docker images.
func NewManagerDocker() (mgr IManager, err error) {
	if !IsDockerAvailable() {
		return nil, &ManagerNotAvailableError{dockerCmd}
	}

	return newManagerDocker(dockerCmd, ContainerManagerTypeV.Docker, nil)
}
