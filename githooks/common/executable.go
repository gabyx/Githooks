package common

// IExecutable defines the interface for a general executable.
type IExecutable interface {
	GetCommand() string
	GetArgs(args ...string) []string

	GetEnvironment() []string

	ResolveExitCode(exitCode int) string
	ApplyEnvironmentToArgs(env []string)
}

// Executable contains the data to a script/executable file.
type Executable struct {
	// The absolute path of the hook script/executable.
	Cmd string

	// Arguments to the hook script/executable.
	Args []string
	Env  []string
}

// GetCommand gets the first command.
func (e *Executable) GetCommand() string {
	return e.Cmd
}

// GetArgs gets all args.
func (e *Executable) GetArgs(args ...string) []string {
	return append(e.Args, args...)
}

// GetEnvironment gets all environment variables.
func (e *Executable) GetEnvironment() []string {
	return e.Env
}

// ApplyEnvironmentToArgs applies env. variables to arguments.
func (e *Executable) ApplyEnvironmentToArgs(env []string) {
	// Dont to anything, since normal command dont need this.
}

// GetExitCodeHelp gets help for any non-zero exit code if needed.
func (e *Executable) ResolveExitCode(exitCode int) string {
	// Not needed for normal commands.
	return ""
}
