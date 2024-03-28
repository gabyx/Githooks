package container

// EnvVariableContainerRunConfigFile is the YAML file which is used
// for all run invocations
// of containerized hooks. Its enables :
//   - to set custom additional arguments to the container run invocation, e.g.
//     mount special volumes or set special environment variables needed in CI.
//   - override workspace path (`/mnt/workspace`) in the container.
//   - override shared repository path `/mnt/shared` in the container.
//
// For example a file:
//
// ```yaml
//
//		version: 1
//		workspace-dir-dest: /builds/a/b/c
//		shared-dir-dest: /builds/.githooks/shared
//	 auto-mount-workspace: false
//	 auto-mount-shared: false
//	 args: [ "--volumes-from", "123455" ]
//
// ```
// Would mount the volumes from container `123455` (podman) and
// use the workspace dir `builds/a/b/c` and `/builds/.githooks/shared`.
const EnvVariableContainerRunConfigFile = "GITHOOKS_CONTAINER_RUN_CONFIG_FILE"
