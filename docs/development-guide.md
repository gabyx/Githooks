# Development Guide

We use a Nix DevShell to have the toolchain ready to develop in this repository,
see
[installation reference](https://swissdatasciencecenter.github.io/best-practice-documentation/docs/dev-enablement/nix-and-nixos#installing-nix)

To enter the DevShell, one can run

```bash
just develop
```

or

```bash
nix develop ".#default"

```

## Building

```shell
just build [--help]
```

builds the executable into `githooks/bin`.

## Linting

```shell
just lint
# or
just lint-fix
```

runs all linting inside this repository containerized.

## Tests and Debugging

The integration tests in [`tests`](../tests) are containerized (legacy
decision). To run the tests you can use:

Running the integration tests you can use `just test *args` which runs

```bash
tests/test-alpine.sh # and other 'test-XXXX.sh' files...
```

Run certain tests only (see `just list-tests`):

```bash
tests/test-alpine.sh --seq {001..120}
tests/test-alpine.sh --seq 065
```

### Debugging in the Dev Container

There is a docker development container for debugging purposes in
`.devcontainer`. VS Code can be launched in this remote docker container with
the extension `ms-vscode-remote.remote-containers`. Use
`Remote-Containers: Open Workspace in Container...` and
`Remote-Containers: Rebuild Container`.

Once in the development container: You can launch the VS Code tasks:

- `[Dev Container] go-delve-installer`
- etc...

which will start the `delve` debugger headless as a server in a terminal. You
can then attach to the debug server with the debug configuration
`Debug Go [remote delve]`. Set breakpoints in the source code to trigger them.
