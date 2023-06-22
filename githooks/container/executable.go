package container

// ContainerizedExecutable contains the data to a script/executable file.
type ContainerizedExecutable struct {
	containerType ContainerManagerType
	usedVolumes   bool

	Cmd string // The command.

	ArgsPre  []string // The pre arguments.
	ArgsEnv  []string // Arguments which set environment values inside the containerized run.
	ArgsPost []string // The post arguments.
}

// GetCommand gets the first command.
func (e *ContainerizedExecutable) GetCommand() string {
	return e.Cmd
}

// GetArgs gets all args.
func (e *ContainerizedExecutable) GetArgs(args ...string) []string {
	var a []string
	a = append(a, e.ArgsPre...)
	a = append(a, e.ArgsEnv...)
	a = append(a, e.ArgsPost...)

	return a
}

// GetEnvironment gets all environment variables.
func (e *ContainerizedExecutable) GetEnvironment() []string {
	return nil
}

// ApplyEnvironmentToArgs applies all environment variables `env` to the arguments of
// the call to be able to forward them into the container.
func (e *ContainerizedExecutable) ApplyEnvironmentToArgs(env []string) {
	if e.containerType == ContainerManagerTypeV.Docker {
		for i := range env {
			e.ArgsEnv = append(e.ArgsEnv, "-e", env[i])
		}
	} else {
		panic("Not implemented.")
	}
}

// GetExitCodeHelp gets help for any non-zero exit code if needed.
func (e *ContainerizedExecutable) ResolveExitCode(exitCode int) string {
	if e.containerType == ContainerManagerTypeV.Docker {
		switch exitCode {
		case 125: // nolint: gomnd
			return "The docker daemon reported an error.\n" +
				"Note: If you are inside a container ALREADY and want\n" +
				"to run hooks containerized (docker-in-docker) you can ONLY do\n" +
				"this by specifying host-machine paths (or a container volume) \n" +
				"for two locations:\n\n" +
				" - path (or container volume) and relative base path pointing to the \n" +
				"   workspace repository on the host machine where Githooks runs in,\n\n" +
				" - path (or container volume) pointing to the shared hooks \n" +
				"   location on the host machine, e.g `~/.githooks/shared`.\n\n" +
				"Check the Githooks manual for instructions on docker-in-docker."
		case 126: // nolint: gomnd
			return "Docker command could not be invoked (permission problem?)."
		case 127: // nolint: gomnd
			return "Command could not be found."
		}
	}

	return ""
}
