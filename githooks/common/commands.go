package common

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"

	strs "github.com/gabyx/githooks/githooks/strings"
)

// CmdContext defines the command context to execute commands.
type CmdContext struct {
	baseCmd      string
	cwd          string
	env          []string
	captureError bool
}

type CmdContextBuilder struct {
	cmdCtx CmdContext
}

func NewCommandCtxBuilder() *CmdContextBuilder {
	return &CmdContextBuilder{cmdCtx: CmdContext{baseCmd: "", cwd: "", env: nil}}
}

func (c *CmdContextBuilder) Build() CmdContext {
	return c.cmdCtx
}

// SetEnv sets the environment.
func (c *CmdContextBuilder) FromCtx(cmdCtx CmdContext) *CmdContextBuilder {
	c.cmdCtx.baseCmd = cmdCtx.baseCmd
	c.cmdCtx.cwd = cmdCtx.cwd
	c.cmdCtx.env = cmdCtx.env
	c.cmdCtx.captureError = cmdCtx.captureError

	return c
}

// SetEnv sets the environment.
func (c *CmdContextBuilder) SetEnv(env []string) *CmdContextBuilder {
	c.cmdCtx.env = env

	return c
}

// SetBaseCmd sets the base command.
func (c *CmdContextBuilder) SetBaseCmd(cmd string) *CmdContextBuilder {
	c.cmdCtx.baseCmd = cmd

	return c
}

// SetCwd sets the working dir.
func (c *CmdContextBuilder) SetCwd(cwd string) *CmdContextBuilder {
	c.cmdCtx.cwd = cwd

	return c
}

// EnableCaptureError enables capturing the `stderr`.
func (c *CmdContextBuilder) EnableCaptureError() *CmdContextBuilder {
	c.cmdCtx.captureError = true

	return c
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
	if strs.IsEmpty(out) {
		return nil, err
	}

	return strs.SplitLines(out), err
}

// Get executes a command and gets the stdout.
func (c *CmdContext) Get(args ...string) (string, error) {
	cmd := exec.Command(c.baseCmd, args...)
	cmd.Dir = c.cwd
	cmd.Env = c.env

	var buf bytes.Buffer
	if c.captureError {
		cmd.Stderr = &buf
	}
	stdout, err := cmd.Output()

	if err != nil {
		var errS string
		exitErr := &exec.ExitError{}

		if c.captureError {
			errS = buf.String()
		} else if errors.As(err, &exitErr) {
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
	cmd.Env = c.env

	stdout, err := cmd.CombinedOutput()

	if err != nil {
		var errS string
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
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
	cmd.Env = c.env

	var buf bytes.Buffer
	if c.captureError {
		cmd.Stderr = &buf
	}

	err = cmd.Run()

	if err != nil {
		var errS string
		if c.captureError {
			errS = buf.String()
		} else {
			exitErr := &exec.ExitError{}
			if errors.As(err, &exitErr) {
				errS = string(exitErr.Stderr)
			}
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
	cmd.Env = c.env

	err := cmd.Run()

	if err == nil {
		return 0, nil
	}

	t := &exec.ExitError{}
	if errors.As(err, &t) {
		return t.ExitCode(), nil
	}

	return -1, CombineErrors(
		ErrorF("Could not get exit status of '%s %q' [cwd: '%s', env: %q].",
			c.baseCmd, args, cmd.Dir, cmd.Env), err)
}

// CheckPiped checks if a command executed successfully.
func (c *CmdContext) CheckPiped(args ...string) (err error) {
	cmd := exec.Command(c.baseCmd, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = c.cwd
	cmd.Env = c.env

	buf := bytes.NewBuffer(nil)
	if c.captureError {
		cmd.Stderr = buf
	}

	err = cmd.Run()

	if err != nil {
		var errS string
		if c.captureError {
			errS = buf.String()
		} else {
			exitErr := &exec.ExitError{}
			if errors.As(err, &exitErr) {
				errS = string(exitErr.Stderr)
			}
		}

		err = CombineErrors(
			ErrorF("Command failed: '%s %q' [cwd: '%s', env: %q, err: '%s'].",
				c.baseCmd, args, cmd.Dir, cmd.Env, errS), err)
	}

	return
}
