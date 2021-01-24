package common

// IExecutable defines the interface for a general executable.
type IExecutable interface {
	GetCommand() string
	GetArgs(args ...string) []string
}

// Executable contains the data to a script/executable file.
type Executable struct {
	// The absolute path of the hook script/executable.
	Cmd string

	// First arguments to the hook script/executable.
	Args []string
}

// GetCommand gets the first command.
func (e *Executable) GetCommand() string {
	return e.Cmd
}

// GetArgs gets all args.
func (e *Executable) GetArgs(args ...string) []string {
	return append(e.Args, args...)
}
