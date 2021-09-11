package builder

import (
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/hashicorp/go-version"
)

var relPathGoSrc = "githooks"
var goVersionMin = "1.16.0"
var versionRe = regexp.MustCompile(`\d+\.\d+(\.\d+)?`)

func findGoExec(cwd string) (cm.CmdContext, error) {

	check := func(gox cm.CmdContext) error {

		verS, err := gox.Get("version")
		if err != nil {
			return cm.ErrorF(
				"Executable '%s' is not found.",
				gox.GetBaseCmd())
		}

		ver := versionRe.FindString(verS)
		if strs.IsEmpty(ver) {
			return cm.ErrorF(
				"Executable version of '%s' cannot be matched.",
				gox.GetBaseCmd())
		}

		verMin, err := version.NewVersion(goVersionMin)
		cm.DebugAssert(err == nil, "Wrong version.")

		verCurr, err := version.NewVersion(ver)
		if err != nil {
			return cm.ErrorF(
				"Executable version '%s' of '%s' cannot be parsed.",
				ver, gox.GetBaseCmd())
		}

		if verCurr.LessThan(verMin) {
			return cm.ErrorF(
				"Executable version of '%s' is '%s' -> min. required is '%s'.",
				gox.GetBaseCmd(), ver, goVersionMin)
		}

		return nil
	}

	var gox cm.CmdContext
	var err error

	// Check from config.
	goExec := git.Ctx().GetConfig(hooks.GitCKGoExecutable, git.GlobalScope)
	if strs.IsNotEmpty(goExec) && cm.IsFile(goExec) {
		gox = cm.NewCommandCtx(goExec, cwd, nil)

		e := check(gox)
		if e == nil {
			return gox, nil
		}
		err = cm.CombineErrors(err, e)
	}

	// Check globally in path.
	gox = cm.NewCommandCtx("go", cwd, nil)
	e := check(gox)
	if e == nil {
		return gox, nil
	}

	return cm.CmdContext{}, cm.CombineErrors(err, e)
}

// Build compiles this repos executable with Go and reports
// the output binary directory where all built binaries reside.
func Build(gitx *git.Context, buildTags []string) (string, error) {

	repoPath := gitx.GetCwd()
	goSrc := path.Join(repoPath, relPathGoSrc)
	if !cm.IsDirectory(goSrc) {
		return "", cm.ErrorF("Source directors '%s' is not existing.", goSrc)
	}

	goPath := path.Join(repoPath, relPathGoSrc, ".go")
	goBinPath := path.Join(repoPath, relPathGoSrc, "bin")

	// Find the go executable
	gox, err := findGoExec(goSrc)
	if err != nil {
		return goBinPath,
			cm.CombineErrors(
				cm.Error("Could not find a suitable 'go' executable."),
				err)
	}

	// Build it.
	e1 := os.RemoveAll(goPath)
	e2 := os.RemoveAll(goBinPath)
	if e1 != nil || e2 != nil {
		return goBinPath, cm.Error("Could not remove temporary build files.")
	}

	// Modify environment for compile.
	gox.Env = strs.Filter(os.Environ(), func(s string) bool {
		return !strings.Contains(s, "GOBIN") &&
			!strings.Contains(s, "GOPATH")
	})

	gox.Env = append(gox.Env,
		strs.Fmt("GOBIN=%s", goBinPath),
		strs.Fmt("GOPATH=%s", goPath))

	// Initialize modules.
	vendorCmd := []string{"mod", "vendor"}
	out, err := gox.GetCombined(vendorCmd...)
	if err != nil {
		return goBinPath,
			cm.ErrorF("Module vendor command failed:\n'%s %q'\nOutput:\n%s",
				gox.GetBaseCmd(), vendorCmd, out)
	}

	// Genereate everything.
	generateCmd := []string{"generate", "-mod=vendor", "./..."}
	out, err = gox.GetCombined(generateCmd...)
	if err != nil {
		return goBinPath,
			cm.ErrorF("Generate command failed:\n'%s %q'\nOutput:\n%s",
				gox.GetBaseCmd(), generateCmd, out)
	}

	// Compile everything.
	cmd := []string{"install", "-mod=vendor"}

	if runtime.GOOS == cm.WindowsOsName {
		buildTags = append(buildTags, cm.WindowsOsName)
	}

	if len(buildTags) != 0 {
		cmd = append(cmd, "-tags", strings.Join(buildTags, ","))
	}

	cmd = append(cmd, "./...")
	out, err = gox.GetCombined(cmd...)

	if err != nil {
		return goBinPath,
			cm.ErrorF("Compile command failed:\n'%s %q'\nOutput:\n%s",
				gox.GetBaseCmd(), cmd, out)
	}

	return goBinPath, nil
}
