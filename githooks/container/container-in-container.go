package container

// EnvVariableContainerWorkspaceHostPath specifies the source mount path (can be a container volume too)
// on the host machine where the Git repository is located in which Githooks works on.
// Normally Githooks uses a bind mount, but for docker-in-docker that does not work
// then we need these variables.
// Example: `~/work/myproject`.
const EnvVariableContainerWorkspaceHostPath = "GITHOOKS_CONTAINER_WORKSPACE_HOST_PATH"

// EnvVariableContainerWorkspaceBasePath specifies a relative path to the host path
// above. Normally empty.
// The variable can contain `${repository-dir-name}` which is replaced by
// the current base name of the repository where Githooks runs.
// Example: `repos/${repository-dir-name}` (Git repo relative to `EnvVariableContainerWorkspaceHostPath`).
const EnvVariableContainerWorkspaceBasePath = "GITHOOKS_CONTAINER_WORKSPACE_BASE_PATH"

// EnvVariableContainerSharedHostPath specifies the source mount path (can be a container volume too)
// on the host machine where the shared hook repositories are located.
// Normally Githooks uses a bind mount, but for docker-in-docker that does not work
// and this variable must be set if shared hooks are needed.
// It makes sense to mount the host `~/.githooks/shared` path directly into
// container at the same place such that they are in sync with what the containerized hooks
// when this variable is set to e.g. `~/.githooks/shared`.
const EnvVariableContainerSharedHostPath = "GITHOOKS_CONTAINER_SHARED_HOST_PATH"
