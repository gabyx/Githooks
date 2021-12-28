package common

import (
	"os"
	"os/exec"
	"strings"

	strs "github.com/gabyx/githooks/githooks/strings"
)

// CmdContext defines the command context to execute commands.
type CmdContext struct {
	Env []string

	baseCmd string
	cwd     string
}

// NewCtx creates a new `CmdContext`.
func NewCommandCtx(baseCmd string, cwd string, env []string) CmdContext {
	return CmdContext{baseCmd: baseCmd, cwd: cwd, Env: env}
}

// GetCwd returns the working directory.
func (c *CmdContext) GetCwd() string {
	return c.cwd
}

// GetBaseCmd returns the base command.
func (c *CmdContext) GetBaseCmd() string {
	return c.baseCmd
}

// GetSplit executes a command and splits the output by newlines.
func (c *CmdContext) GetSplit(args ...string) ([]string, error) {
	out, err := c.Get(args...)

	return strs.SplitLines(out), err
}

// Get executes a command and gets the stdout.
func (c *CmdContext) Get(args ...string) (string, error) {
	cmd := exec.Command(c.baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.Env
	stdout, err := cmd.Output()

	if err != nil {
		var errS string
		if exitErr, ok := err.(*exec.ExitError); ok {
			errS = string(exitErr.Stderr)
		}
		err = CombineErrors(
			ErrorF("Command failed: '%s %q' [cwd: '%s', env: %q, err: '%s'].",
				c.baseCmd, args, cmd.Dir, cmd.Env, errS), err)
	}

	return strings.TrimSpace(string(stdout)), err
}

// GetCombined executes a command and gets the combined stdout and stderr.
func (c *CmdContext) GetCombined(args ...string) (string, error) {
	cmd := exec.Command(c.baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.Env

	stdout, err := cmd.CombinedOutput()

	if err != nil {
		var errS string
		if exitErr, ok := err.(*exec.ExitError); ok {
			errS = string(exitErr.Stderr)
		}
		err = CombineErrors(
			ErrorF("Command failed: '%s %q' [cwd: '%s', env: %q, err: '%s'].",
				c.baseCmd, args, cmd.Dir, cmd.Env, errS), err)
	}

	return strings.TrimSpace(string(stdout)), err
}

// Check checks if a command executed successfully.
func (c *CmdContext) Check(args ...string) (err error) {
	cmd := exec.Command(c.baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.Env

	err = cmd.Run()

	if err != nil {
		var errS string
		if exitErr, ok := err.(*exec.ExitError); ok {
			errS = string(exitErr.Stderr)
		}

		err = CombineErrors(
			ErrorF("Command failed: '%s %q' [cwd: '%s', env: %q, err: '%s'].",
				c.baseCmd, args, cmd.Dir, cmd.Env, errS), err)
	}

	return
}

// GetExitCode get the exit code of the command.
func (c *CmdContext) GetExitCode(args ...string) (int, error) {
	cmd := exec.Command(c.baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.Env

	err := cmd.Run()

	if err == nil {
		return 0, nil
	}

	if t, ok := err.(*exec.ExitError); ok {
		return t.ExitCode(), nil
	}

	return -1, CombineErrors(
		ErrorF("Could get exit status of '%s %q' [cwd: '%s', env: %q].",
			c.baseCmd, args, cmd.Dir, cmd.Env), err)
}

// CheckPiped checks if a command executed successfully.
func (c *CmdContext) CheckPiped(args ...string) (err error) {
	cmd := exec.Command(c.baseCmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = c.cwd
	cmd.Env = c.Env

	err = cmd.Run()

	if err != nil {
		var errS string
		if exitErr, ok := err.(*exec.ExitError); ok {
			errS = string(exitErr.Stderr)
		}

		err = CombineErrors(
			ErrorF("Command failed: '%s %q' [cwd: '%s', env: %q, err: '%s'].",
				c.baseCmd, args, cmd.Dir, cmd.Env, errS), err)
	}

	return
}
