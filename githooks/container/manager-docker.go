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

// NewHookRunExec runs a hook over a container.
func (m *ManagerDocker) NewHookRunExec(
	ref string,
	workspaceDir string,
	hookRepoDir string,
	hookExec cm.IExecutable,
) (cm.IExecutable, error) {
	containerExec := cm.Executable{}

	containerExec.Cmd = dockerCmd

	mntWorkspace := "/mnt/workspace"
	mntHookRepo := "/mnt/shared"

	cmd := hookExec.GetCommand()

	// Resolve commands with path separators which are
	// relative paths relative to the `rootDir`.
	// e.g `dist/custom.exe` -> `rootDir/dist/custom.exe`
	if strings.ContainsAny(hookExec.GetCommand(), "/\\") {
		if runtime.GOOS == cm.WindowsOsName {
			cmd = filepath.ToSlash(cmd)
		}

		if !filepath.IsAbs(cmd) {
			cmd = path.Join(mntHookRepo, cmd)
		}
	}

	cm.DebugAssertF(!strings.Contains(workspaceDir, "\\"),
		"No forward slashes should be passed in here '%s'.", workspaceDir)
	cm.DebugAssertF(!strings.Contains(hookRepoDir, "\\"),
		"No forward slashes should be passed in here '%s'.", hookRepoDir)

	containerExec.Args = []string{
		"run",
		"--rm",
		"-v",
		strs.Fmt("%v:%v", workspaceDir, mntWorkspace),
		"-v",
		strs.Fmt("%v:%v:ro", hookRepoDir, mntHookRepo),
		"-w", mntWorkspace,
		"-e", "GITHOOKS_CONTAINER_RUN=true",
	}

	if runtime.GOOS != cm.WindowsOsName && runtime.GOOS != "darwin" {
		containerExec.Args = append(containerExec.Args, "--user", strs.Fmt("%v:%v", m.uid, m.gid))
	}

	// Re-export env variables (does not contain general environment).
	for _, envKeyVar := range hookExec.GetEnvironment() {
		containerExec.Args = append(containerExec.Args, "-e", envKeyVar)
	}

	containerExec.Args = append(containerExec.Args, ref, cmd)
	containerExec.Args = append(containerExec.Args, hookExec.GetArgs()...)

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
