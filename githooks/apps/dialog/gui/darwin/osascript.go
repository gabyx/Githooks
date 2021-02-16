//go:generate go run -mod=vendor ../../tools/embed-files.go

package gui

import (
	"context"
	"gabyx/githooks/apps/dialog/build"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"os/exec"
	"path"
	"strings"
	"text/template"
)

func RunOSAScript(ctx context.Context, script string, data interface{}, workingDir string) ([]byte, error) {
	var buf strings.Builder

	tmpl, err := build.Asset(path.Join("gui/darwin/osascripts", script+".js.tmpl"))
	cm.AssertNoErrorPanic(err, "Template not embedded.")
	template, err := template.New("").Funcs(templateFuncs).Parse(string(tmpl))
	cm.AssertNoErrorPanic(err, "Template '%s' invalid.", script)

	err = template.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	script = buf.String()
	lang := "JavaScript"

	var cmd *exec.Cmd
	if ctx != nil {
		cmd = exec.CommandContext(ctx, "osascript", "-l", lang)

	} else {
		cmd = exec.Command("osascript", "-l", lang)
	}

	if strs.IsNotEmpty(workingDir) {
		cmd.Dir = workingDir
	}

	cmd.Stdin = strings.NewReader(script)

	out, err := cmd.Output()
	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	return out, err
}
