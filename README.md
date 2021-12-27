<img src="docs/githooks-logo.svg" style="margin-left: 20pt" align="right">
<h1>Githooks</h1>

[![Build Status](https://travis-ci.org/gabyx/githooks.svg?branch=main)](https://travis-ci.org/gabyx/githooks)
[![Coverage Status](https://coveralls.io/repos/github/gabyx/Githooks/badge.svg?branch=main)](https://coveralls.io/github/gabyx/Githooks?branch=main)
[![goreleaser](https://github.com/gabyx/Githooks/actions/workflows/release.yml/badge.svg?branch=prepare-v2.0.3)](https://github.com/gabyx/Githooks/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gabyx/githooks)](https://goreportcard.com/report/github.com/gabyx/githooks)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/nlohmann/json/master/LICENSE.MIT)
[![GitHub Releases](https://img.shields.io/github/release/gabyx/githooks.svg)](https://github.com/gabyx/githooks/releases)
![Git Version](https://img.shields.io/badge/Git-%E2%89%A5v.2.28.0-blue)
![Go Version](https://img.shields.io/badge/Go-1.17-blue)
![OS](https://img.shields.io/badge/OS-linux,%20macOs,%20Windows-blue)

A **platform-independend hooks manager** written in Go to support shared hook
repositories and per-repository
[Git hooks](https://git-scm.com/docs/cli/githooks), checked into the working
repository. This implementation is the Go port and successor of the
[original impementation](https://github.com/rycus86/githooks) (see
[Migration](#migrating)).

To make this work, the installer creates run-wrappers for Githooks that are
installed into the `.git/hooks` folders automatically on `git init` and
`git clone`. There's more [to the story though](#templates-or-global-hooks).
When one of the Githooks run-wrappers executes, Githooks starts up and tries to
find matching hooks in the `.githooks` directory under the project root, and
invoke them one-by-one. Also it searches for hooks in configured shared hook
repositories.

**This Git hook manager supports:**

- Running repository checked-in hooks.
- Running shared hooks from other Git repositories (with auto-update).
- Git LFS support.
- Command line interface.
- Fast execution due to compiled Go executable.
- Fast parallel execution over threadpool.
- Ignoring non-shared and shared hooks with patterns.
- Automatic Githooks updates: Fully configurable for your own company by
  url/branch and deploy settings.
- **Bonus:** [Platform-independent dialog tool](#dialog-tool) for user prompts
  inside your own hooks.

<details>
<summary><b>Table of Content (click to expand)</b></summary>
<!-- TOC -->

- [Layout and Options](#layout-and-options)
- [Execution](#execution)
  - [Hook Run Configuration](#hook-run-configuration)
  - [Exported Environment Variables](#exported-environment-variables)
  - [Parallel Execution](#parallel-execution)
- [Supported Hooks](#supported-hooks)
- [Git Large File Storage (Git LFS) Support](#git-large-file-storage-git-lfs-support)
- [Shared hook repositories](#shared-hook-repositories)
  - [Global Configuration](#global-configuration)
  - [Local Configuration](#local-configuration)
  - [Repository Configuration](#repository-configuration)
  - [Supported URLS](#supported-urls)
  - [Skip Non-Existing Shared Hooks](#skip-non-existing-shared-hooks)
- [Layout of Shared Hook Repositories](#layout-of-shared-hook-repositories)
  - [Shared Repository Namespace](#shared-repository-namespace)
- [Ignoring Hooks and Files](#ignoring-hooks-and-files)
- [Trusting Hooks](#trusting-hooks)
- [Disabling Githooks](#disabling-githooks)
- [Environment Variables](#environment-variables)
- [Log & Traces](#log--traces)
- [Installing or Removing Run-Wrappers](#installing-or-removing-run-wrappers)
- [User Prompts](#user-prompts)
- [Installation](#installation)
  - [Normal Installation](#normal-installation)
  - [Install Mode: Centralized Hooks](#install-mode-centralized-hooks)
  - [Install from different URL and Branch](#install-from-different-url-and-branch)
  - [No Installation](#no-installation)
  - [Non-Interactive Installation](#non-interactive-installation)
  - [Install on the Server](#install-on-the-server)
    - [Setup for Bare Repositories](#setup-for-bare-repositories)
  - [Templates or Global Hooks](#templates-or-global-hooks)
    - [Template Folder (`init.templateDir`)](#template-folder-inittemplatedir)
    - [Global Hooks Location (`core.hooksPath`)](#global-hooks-location-corehookspath)
  - [Updates](#updates)
    - [Update Mechanics](#update-mechanics)
- [Uninstalling](#uninstalling)
- [YAML Specifications](#yaml-specifications)
- [Migration](#migration)
- [Dialog Tool](#dialog-tool)
  - [Build From Source](#build-from-source)
  - [Dependencies](#dependencies)
- [Tests and Debugging](#tests-and-debugging)
  - [Debugging in the Dev Container](#debugging-in-the-dev-container)
  - [Todos:](#todos)
- [Changelog](#changelog)
  - [Version v2.x.x](#version-v2xx)
- [FAQ](#faq)
- [Acknowledgements](#acknowledgements)
- [Authors](#authors)
- [Support & Donation](#support--donation)
- [License](#license)

<!-- /TOC -->
</details>

## Layout and Options

Take this snippet of a Git repository layout as an example:

```bash
/
├── .githooks/
│    ├── commit-msg/          # All commit-msg hooks.
│    │    ├── validate        # Normal hook script.
│    │    └── add-text        # Normal hook script.
│    │
│    ├── pre-commit/          # All pre-commit hooks.
│    │    ├── .ignore.yaml    # Ignores relative to 'pre-commit' folder.
│    │    ├── 01-validate     # Normal hook script.
│    │    ├── 02-lint         # Normal hook script.
│    │    ├── 03-test.yaml    # Hook run configuration.
│    │    ├── docs.md         # Ignored in '.ignore.yaml'.
│    │    └── final/          # Batch folder 'final' which runs all in parallel.
│    │        ├── 01-validate # Normal hook script.
│    │        └── 02-upload   # Normal hook script.
│    │
│    ├── post-checkout/       # All post-checkout hooks.
│    │   ├── .all-parallel    # All hooks in this folder run in parallel.
│    │   └── ...
│    ├── ...
│    ├── .ignore.yaml         # Main ignores.
│    ├── .shared.yaml         # Shared hook configuration.
│    └── .lfs-required        # LFS is required.
└── ...
```

All hooks to be executed live under the `.githooks` top-level folder, that
should be checked into the repository. Inside, we can have directories with the
name of the hook (like `commit-msg` and `pre-commit` above), or a file matching
the hook name (like `post-checkout` in the example). The filenames in the
directory do not matter, but the ones starting with a `.` (dotfiles) will be
excluded by default. All others are executed in lexical order according to the
Go function [`Walk`](https://golang.org/pkg/path/filepath/#Walk). rules.
Subfolders as e.g. `final` get treated as parallel batch and all hooks inside
are by default executed in parallel over the thread pool. See
[Parallel Execution](#parallel-execution) for details.

You can use the [command line helper](docs/cli/git_hooks.md) (a globally
configured Git alias `alias.hooks`), that is `git hooks list`, to list all hooks
and their current state that apply to the current repository. For this
repository this [looks like the following.](docs/githooks-list.png)

## Execution

If a file is executable, it is directly invoked, otherwise it is interpreted
with the `sh` shell. On Windows that mostly means dispatching to the `bash.exe`
from [https://gitforwindows.org](https://gitforwindows.org).

**All parameters and standard input** are forwarded from Git to the hooks. The
standard output and standard error of any hook which Githooks runs is captured
**together**<span id="a1">[<sup>1</sup>](#1)</span> and printed to the standard
error stream which might or might not get read by Git itself (e.g. `pre-push`).

Hooks can also be specified by a run configuration in a corresponding YAML file,
see [Hook Run Configuration](#hook-run-configuration).

Hooks related to `commit` events (where it makes sense, not `post-commit`) will
also have a `${STAGED_FILES}` environment variable set, i.e. the list of staged
and changed files according to
`git diff --cached --diff-filter=ACMR --name-only`. File paths are separated by
a newline `\n`. If you want to iterate in a shell script over them, and expect
spaces in paths, you might want to set the `IFS` like this:

```shell
IFS="
"
for STAGED in ${STAGED_FILES}; do
    ...
done
```

The `ACMR` filter in the `git diff` will include staged files that are added,
copied, modified or renamed.

**<span id="1"><sup>1</sup></span>[⏎](#a1) Note:** This caveat is basically
there because standard output and error might get interleaved badly and so far
no solution to this small problem has been tackled yet. It is far better to
output both streams in the correct order, and therefore send it to the error
stream because that will not conflict in anyway with Git (see
[fsmonitor-watchman](https://git-scm.com/docs/githooks#_fsmonitor_watchman),
unsupported right now.). If that poses a real problem for you, open an issue.

### Hook Run Configuration

Each supported hook can also be specified by a configuration file
`<hookName>.yaml` where `<hookName>` is any
[supported hook name](#supported-hooks). An example might look like the
following:

```yaml
# The command to run.
# - if it contains path separators and is relative, it its evaluated relative to
#   the worktree of the repository where this config resides.
cmd: "dist/command-of-${env:USER}.exe"

# The arguments given to `cmd`.
args:
  - "-s"
  - "--all"
  - "${env:GPG_PUBLIC_KEY}"
  - "--test ${git-l:my-local-git-config-var}"

# If you want to make sure your file is not
# treated always as the newest version. Fix the version by:
version: 1
```

All additional arguments given by Git to `<hookName>` will be appended last onto
`args`. All environment and Git config variables in `args` and `cmd` are
substituted with the following syntax:

- `${env:VAR}` : An environment variable `VAR`.
- `${git:VAR}` : A Git config variable `VAR` which corresponds to
  `git config 'VAR'`.
- `${git-l:VAR}` : A Git config variable `VAR` which corresponds to
  `git config --local 'VAR'`.
- `${git-g:VAR}` : A Git config variable `VAR` which corresponds to
  `git config --global 'VAR'`.
- `${git-s:VAR}` : A Git config variable `VAR` which corresponds to
  `git config --system 'VAR'`.

Not existing environment variables or Git config variables are replaced with the
empty string by default. If you use `${!...:VAR}` (e.g `${!git-s:VAR }`) it will
trigger an error and fail the hook if the variable `VAR` is not found. Escaping
the above syntax works with `\${...}`.

**Sidenote**: You might wonder why this configuration is not gathered in one
single YAML file for all hooks. The reason is that each hook invocation by Git
is separate. Avoiding reading this total file several times needs time and since
we want speed and only an opt-in solution this is avoided.

### Exported Environment Variables

Githooks defines the following environment variables:

- `STAGED_FILES` : All staged files, applicable to hooks: `pre-commit`,
  `prepare-commit-msg` and `commit-msg`.
- `GITHOOKS_OS` : The runtime operating system, applicable to all hooks.
  [See this list](https://github.com/golang/go/blob/master/src/go/build/syslist.go).
- `GITHOOKS_ARCH` : The runtime operating architecture, applicable to all hooks.
  [See this list](https://github.com/golang/go/blob/master/src/go/build/syslist.go).

### Parallel Execution

As in the [example](#layout-and-options), all discovered hooks in subfolders
`<batchName>`, e.g. `<repoPath>/<hooksDir>/<hookName>/<batchName>/*` where
`<hooksDir>` is either

- `.githooks` for repository checked-in hooks or
- `githooks`, `.githooks` or `.` for shared repository hooks,

are assigned the same batch name `<batchName>` and processed in parallel. Each
batch is a synchronisation point and starts after the one before has finished.
The threadpool uses by default as many threads as cores on the system. The
number of threads can be controlled by the Git configuration variable
`githooks.numThreads` set anywhere, e.g. in the local or global Git
configuration.

If you place a file `.all-parallel` inside `<hooksDir>/<hookName>`, all
discovered hooks inside `<hooksDir>/<hookName>` are assigned to the same batch
name `all` resulting in executing all hooks in one parallel batch.

You can inspect the computed batch name by running
[`git hooks list --batch-name`](/docs/cli/git_hooks_list.md).

## Supported Hooks

The supported hooks are listed below. Refer to the
[Git documentation](https://git-scm.com/docs/cli/githooks) for information on
what they do and what parameters they receive.

It is receommended to use `--maintained-hooks` options during install
([1](#normal-installation), [2](#installing-or-removing-run-wrappers)) to only
select the hooks which are really needed, since executing the Githooks manager
for all hooks might slow down Git operations (especially for
`reference-transaction`).

- `applypatch-msg`
- `pre-applypatch`
- `post-applypatch`
- `pre-commit`
- `pre-merge-commit`
- `prepare-commit-msg`
- `commit-msg`
- `post-commit`
- `pre-rebase`
- `post-checkout` (non-zero exit code is wrapped to 1)
- `post-merge`
- `pre-push`
- `pre-receive`
- `update`
- `post-receive`
- `post-update`
- `reference-transaction`
- `push-to-checkout`
- `pre-auto-gc`
- `post-rewrite`
- `sendemail-validate`
- `post-index-change`

The hook `fsmonitor-watchman` is currently not supported. If you have a use-case
for it and want to use it with this tool, please open an issue.

## Git Large File Storage (Git LFS) Support

If the user has installed [Git Large File Storage](https://git-lfs.github.com/)
(`git-lfs`) by calling `git lfs install` globally or locally for a repository
only, `git-lfs` installs 4 hooks when initializing (`git init`) or cloning
(`git clone`) a repository:

- `post-checkout`
- `post-commit`
- `post-merge`
- `pre-push`

Since Githooks overwrites the hooks in `<repoPath>/.git/hooks`, it will also run
all _Git LFS_ hooks internally if the `git-lfs` executable is found on the
system path. You can enforce having `git-lfs` installed on the system by placing
a `<repoPath>/.githooks/.lfs-required` file inside the repository, then if
`git-lfs` is missing, a warning is shown and the hook will exit with code `1`.
For some `post-*` hooks this does not mean that the outcome of the git command
can be influenced even tough the exit code is `1`, for example `post-commit`
hooks can't fail commits. A clone of a repository containing this file might
still work but would issue a warning and exit with code `1`, a push - however -
will fail if `git-lfs` is missing.

It is advisable for repositories using _Git LFS_ to also have a pre-commit hook
(e.g. `examples/lfs/pre-commit`) checked in which enforces a correct
installation of _Git LFS_.

## Shared hook repositories

The hooks are primarily designed to execute programs or scripts in the
`<repoPath>/.githooks` folder of a single repository. However there are
use-cases for common hooks, shared between many repositories with similar
requirements and functionality. For example, you could make sure Python
dependencies are updated on projects that have a `requirements.txt` file, or an
`mvn verify` is executed on `pre-commit` for Maven projects, etc.

For this reason, you can place a `.shared.yaml` file (see
[specs](#yaml-specifications)) inside the `<repoPath>/.githooks` folder, which
can hold a list of repositories which contain common and shared hooks.
Alternatively, you can have shared repositories set by multiple
`githooks.shared` local or global Git configuration variables, and the hooks in
these repositories will execute for all local projects where Githooks is
installed. See [git hooks shared](docs/cli/git_hooks_shared.md) for configuring
all 3 types of shared hooks repositories.

Below are example values for these setting.

### Global Configuration

```shell
$ git config --global --get-all githooks.shared # shared hooks in global config (for all repositories)
https://github.com/shared/hooks-python.git
git@github.com:shared/repo.git@mybranch
```

### Local Configuration

```shell
$ cd myrepo
$ git config --local --get-all githooks.shared # shared hooks in local config (for specific repository)
ssh://user@github.com/shared/special-hooks.git@v3.3.3
/opt/myspecialhooks
```

### Repository Configuration

A example config `<repoPath>/.githooks/shared.yaml` (see
[specs](#yaml-specifications)):

```yaml
version: 1
urls:
  - ssh://user@github.com/shared/special-hooks.git@otherbranch
  - git@github.com:shared/repo.git@mybranch
```

The install script offers to set up shared hooks in the global Git config. but
you can do it any time by changing the global configuration variable.

### Supported URLS

Supported URL for shared hooks are:

- **All URLs [Git supports](https://git-scm.com/docs/cli/git-clone#_git_urls)**
  such as:

  - `ssh://github.com/shared/hooks-maven.git@mybranch` and also the short `scp`
    form `git@github.com:shared/hooks-maven.git`
  - `git://user@github.com/shared/hooks-python.git`
  - `file:///local/path/to/bare-repo.git@mybranch`

  All URLs can include a tag specification syntax at the end like `...@<tag>`,
  where `<tag>` is a Git tag, branch or commit hash. The `file://` protocol is
  treated the same as a local path to a bare repository, _see next point_.

- **Local paths** to bare and non-bare repositories such as:

  - `/local/path/to/checkout` (gets used directly)
  - `/local/path/to/bare-repo.git@mybranch` (gets cloned internally)

  Note that relative paths are relative to the path of the repository executing
  the hook. These entries are forbidden for **shared hooks** configured by
  `.githooks/.shared.yaml` per repository because it makes little sense and is a
  security risk.

Shared hooks repositories specified by _URLs_ and _local paths to bare
repository_ will be checked out into the `<installPrefix>/.githooks/shared`
folder (`~/.githooks/shared` by default), and are updated automatically after a
`post-merge` event (typically a `git pull`) on any local repositories. Any other
local path will be used **directly and will not be updated or modified**.
Additionally, the update of shared hook repositories can also be triggered on
other hook names by setting a comma-separated list of additional hook names in
the Git configuration parameter `githooks.sharedHooksUpdateTriggers` on any
configuration level.

You can also manage and update shared hook repositories using the
[`git hooks shared update`](docs/cli/git_hooks_shared.md) command.

### Skip Non-Existing Shared Hooks

**By default, Githooks will fail if any configured shared hooks are not
available and you need to update them by running `git hooks update`**.

By using
[`git hooks config skip-non-existing-shared-hooks --help`](docs/cli/skip-non-existing-shared-hooks.md)
you can disable this behavior locally/globally or by environment variable
`GITHOOKS_SKIP_NON_EXISTING_SHARED_HOOKS` (see
[env. variables](#environment-variables)) which makes Githooks skip non-existing
shared hooks.

## Layout of Shared Hook Repositories

The layout of these shared repositories is the same as above, with the exception
that the hook folders (or files) can be at the project root as well, to avoid
the redundant `.githooks` folder.

If you want the shared hook repository to use Githooks itself (e.g. for
development purposes by using hooks from `<sharedRepo>/.githooks`) you can
furthermore place the _shared_ hooks inside a `<sharedRepo/githooks` subfolder.
In that case the `<sharedRepo>/.githooks` folder is ignored when other users use
this shared repository.

The priority to find hooks in a shared hook repository is as follows: consider
hooks

1. in `<hooksDir> := <sharedRepo>/githooks`, if it does not exist, consider
   hooks
2. in `<hooksDir> := <sharedRepo>/.githooks`, if it does not exist consider
   hooks
3. in `<hooksDir> := <sharedRepo>` as the last fallback.

Each of these directories can be of the same format as the normal `.githooks`
folder in a single repository.

You can get the root directory of a configured shared repository with namespace
`<namespace>` by running `git hooks shared root ns:<namespace>`. This might be
helpful in scripts if you have common shared functionality inside this shared
repository you want to use.

### Shared Repository Namespace

A shared repository can optionally have a namespace associated with it. The name
can be stored in a file `.namespace` in any possible hooks directory
`<hooksDir>` of the shared repository, see
[layout](#layout-of-shared-hook-repositories). The namespace comes into play
when ignoring/disabling certain hooks. See
[ignoring hooks](#ignoring-hooks-and-files). The namespace name must not contain
white spaces (`\s`) or slashes `/`.

The following namespaces names are reserved interally:

- `gh-self` : for hooks in the repository where Githooks runs (if no
  `.namespace` is existing).
- `gh-self-repl` : for original Git hooks which were replaced by Githooks during
  install.

## Ignoring Hooks and Files

The `.ignore.yaml` (see [specs](#yaml-specifications)) files allow excluding
files

- from being treated as hook scripts or
- hooks from beeing run.

You can ignore executing all sorts of hooks per Git repository by specifying
**patterns** or explicit **paths** which match against a hook's (file's)
_namespace path_.

Each hook either in the current repository `<repoPath>/.githooks/...` or inside
a shared hooks repository has a so called _namespace path_.

A _namespace path_ consists of the _name of the hook_ prefixed by a _namespace_
, e.g. :

```
  <namespacePath> := ns:<namespace>/<relPath> = "ns:core-hooks/pre-commit/check-numbers.py"
```

where `<relPath> = pre-commit/check-numbers.py` is the relative path to the
hook. Each shared repository can provide its own
[namespace](#shared-repository-namespace).

A [namespace](#shared-repository-namespace) will be used when the hook belongs
to a shared hook repository and will have a default unique value if it is not
defined. You can inspect all _namespace paths_ by inspecting `ns-path:` in the
output of [git hooks list](docs/cli/git_hooks_list.md) in the current
repository. All ignore entries in `.ignore.yaml` (patterns or paths) will match
against these _namespace paths_.

Disabling works like:

```shell
# Disable certain hooks by a pattern in this repository:
# User ignore pattern stored in `.git/.githooks.ignore.yaml`:
$ git hooks ignore add --pattern "pre-commit/**" # Store: `.git/.githooks.ignore.yaml`:
# or stored inside the repository:
$ git hooks ignore add --repository --pattern "pre-commit/**" # Store: `.githooks/.ignore.yaml`:

# Disable certain shared hooks (with namespace 'my-shared-super-hooks')
# by a glob pattern in this repository:
$ git hooks ignore add --repository --pattern "my-shared-super-hooks://pre-commit/**"
```

In the above [example](#layout-and-options]), one of the `.ignore.yaml` files
should contain a glob pattern `**/*.md` to exclude the `pre-commit/docs.md`
Markdown file. Patterns can contain double star syntax to match multiple
directories, e.g. `**/*.txt` instead of `*.txt`.

The main ignore file `<repoPath>/<hookDir>/.ignore.yaml` applies to all hooks.
Any additional `<repoPath>/<hookDir>/<hookName>/.ignore.yaml` file inside
`<hookDir>` will be accumulated to the main file and patterns not starting with
`ns:` **are made relative to the folder `<hookName>`**. You can also manage
`.ignore.yaml` files using
[`git hooks ignore [add|remove] --help`](docs/cli/git_hooks_ignore.md). Consult
this command documentation for futher information on the pattern syntax.

## Trusting Hooks

To try and make things a little bit more secure, Githooks checks if any new
hooks were added we haven't run before, or if any of the existing ones have
changed. When they have, it will prompt for confirmation (trust prompt) whether
you accept those changes or not, and you can also disable specific hooks to skip
running them until you decide otherwise. The trust prompt is always **fatal**
meaning that failing to answer the prompt, or any other prompt error, will
result in a failing Git hook. To make the `runner` non-interactive, see
[user prompts](#user-prompts). If a hook is still _active and untrusted_ after
the prompt, **Githooks will fail by default**. This is useful to be sure that
all hooks get executed. However, you can disabled this behavior by skipping
active, untrusted hooks with
[`git hooks config skip-untrusted-hooks --enable`](docs/cli/git_hooks_config_skip-untrusted-hooks.md)
or by setting `GITHOOKS_SKIP_UNTRUSTED_HOOKS` (see
[env. variables](#environment-variables)).

The accepted checksums are maintained in the
`<repoPath>/.git/.githooks.checksum` directory, per local repository. You can
however use a global checksum directory setup by making an absolute symbolic
link with name `.githooks.checksum` inside the template directory
(`init.templateDir`) which gets installed in each clone.

If the repository contains a `<repoPath>/.githooks/trust-all` file, it is marked
as a trusted repository. Consult
[`git hooks trust --help`](docs/cli/git_hooks_trust.md). On the first
interaction with hooks, Githooks will ask for confirmation that the user trusts
all existing and future hooks in the repository, and if she does, no more
confirmation prompts will be shown. This can be reverted by running
[`git hooks config trust-all --reset`](docs/cli/git_hooks_config_trust-all.md)
command. This is a per-repository setting. Consult
[`git hooks config trust-all --help`](docs/cli/git_hooks_config_trust-all.md)
for more information.

You can also trust individual hooks by using
[`git hooks trust hooks --help`](docs/cli/git_hooks_trust_hooks.md).

## Disabling Githooks

To disable running any Githooks locally or globally, use the following:

```shell
# Disable Githooks completely for this repository:
$ git hooks disable # Use --reset to undo.
# or
$ git hooks config disable --set # Same thing... Config: `githooks.disable`


# Disable Githooks globally (for all repositories):
$ git hooks disable --global # Use --reset to undo.
# or
$ git hooks config disable --set --global # Same thing... Config: `githooks.disable`
```

Also, as mentioned above, all hook executions can be bypassed with a non-empty
value in the `GITHOOKS_DISABLE` environment variable.

## Environment Variables

All of these environment variables are either defined during Githooks runner
executing or affect its behavior. These should mostly only be used locally and
not globally be defined.

| Environment Variables                          | Effect                                                                                                                    |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `GITHOOKS_OS` (defined)                        | The operating system. <br>See [Exported Environment Variables](#exported-environment-variables).                          |
| `GITHOOKS_ARCH` (defined)                      | The system architecture. <br>See [Exported Environment Variables](#exported-environment-variables).                       |
| `GITHOOKS_DISABLE`                             | If defined, disables running hooks run by Githooks,<br>except `git lfs` and the replaced old hooks.                       |
| `GITHOOKS_RUNNER_TRACE`                        | If defined, enables tracing during <br>Githooks runner execution. A value of `1` enables more output.                     |
| `GITHOOKS_SKIP_NON_EXISTING_SHARED_HOOKS=true` | Skips on `true` and fails on `false` (or empty) for non-existing shared hooks. <br>See [Trusting Hooks](#trusting-hooks). |
| `GITHOOKS_SKIP_UNTRUSTED_HOOKS=true`           | Skips on `true` and fails on `false` (or empty) for untrusted hooks. <br>See [Trusting Hooks](#trusting-hooks).           |

## Log & Traces

You can see how the Githooks `runner` is been called by setting the environment
variable `GITHOOKS_RUNNER_TRACE` to a non empty value.

```shell
GITHOOKS_RUNNER_TRACE=1 git <command> ...
```

## Installing or Removing Run-Wrappers

You can install and uninstall run-wrappers inside a repository with
[`git hooks install`](docs/cli/git_hooks_install.md). or
[`git hooks uninstall`](docs/cli/git_hooks_install.md). This installs and
uninstalls wrappers from `${GIT_DIR}/hooks` as well as sets and unsets local
Githooks-internal Git configuration variables.

To install run-wrappers for only selective hooks, use `--maintained-hooks`, e.g.

```shell
cd repository
git hook install \
    --maintained-hooks "!all, pre-commit, pre-merge-commit, prepare-commit-msg, commit-msg, post-commit" \
    --maintained-hooks "pre-rebase, post-checkout, post-merge, pre-push"
```

**Note:** Git LFS hooks is properly taken care of when `--maintained-hooks` is
used. That is, when you don't select a Git LFS hooks in `--maintained-hooks`,
the missing Git LFS hooks will be installed too.

## User Prompts

Githooks shows user prompts during installation, updating (automatic or manual),
uninstallation and when executing hooks (the `runner` executable).

The `runner` might get executed over a Git GUI or any other environment where no
terminal is available. In this case all user prompts are shown as GUI dialogs
with the included [platform-independent dialog tool](#dialog-tool). The GUI
dialog fallback is currently only enabled for the `runner`.

Githooks distinguishes between _fatal_ and _non-fatal_ prompts.

- A _fatal_ prompt will result in a complete abort if

  - The prompt could not be shown (terminal or GUI dialog).
  - The answer returned by the user is incorrect (terminal only) or the user
    canceled the GUI dialog.

- A _non-fatal_ prompt always has a default answer which is taken in the above
  failing cases and the execution continues. Warning messages might be shown
  however.

The `runner` will show prompts, either in the terminal or as GUI dialog, in the
following cases:

1. **Trust prompt**: The user is required to trust/untrust a new/changed hook:
   **fatal**.
2. **Update prompts**: The user is requested to accept a new update if automatic
   updates are enabled (`git hooks update --enable`): **non-fatal**.
   - Various other prompts when the updater is launched: **non-fatal**.

User prompts during `runner` execution are sometimes not desirable (server
infastructure, docker container, etc...) and need to be disabled. Setting
`git hooks config non-interactive-runner --enable --global` will:

- Take default answers for all **non-fatal** prompts. No warnings are shown.
- Take default answer for a **fatal prompt** if it is configured: The only fatal
  prompt is the **trust prompt** which can be configured to pass by executing
  `git hooks config trust-all --accept`.

## Installation

- [Download the latest release](https://github.com/gabyx/githooks/releases),
  exctract it and execute the installer command by the below instructions.

The installer will:

1. Find out where the Git templates directory is.
   1. From the `$GIT_TEMPLATE_DIR` environment variable.
   2. With the `git config --get init.templateDir` command.
   3. Checking the default `/usr/share/git-core/templates` folder.
   4. Search on the filesystem for matching directories.
   5. Offer to set up a new one, and make it `init.templateDir`.
2. Write all chosen Githooks run-wrappers into the choosen directory:
   - either `init.templateDir` or
   - `core.hooksPath` depending on the install mode `--use-core-hooks-path`.
3. Offer to enable automatic update checks.
4. Offer to find existing Git repositories on the filesystem (disable with
   `--skip-install-into-existing`)
   1. Install run-wrappers into them (`.git/hooks`).
   2. Offer to add an intro README in their `.githooks` folder.
5. Install/update run-wrappers into all registered repositories: Repositories
   using Githooks get registered in the install folders `registered.yaml` file
   on their first hook invocation.
6. Offer to set up shared hook repositories.

### Normal Installation

To install Githooks on your system, simply execute `cli installer`. It will
guide you through the installation process. Check the `cli installer --help` for
available options. Some of them are described below:

Its adviced to only install Githooks for a selection of the supported hooks by
using `--maintained-hooks` as

```shell
cli installer \
    --maintained-hooks "!all, pre-commit, pre-merge-commit, prepare-commit-msg, commit-msg, post-commit" \
    --maintained-hooks "pre-rebase, post-checkout, post-merge, pre-push"
```

This will only support the mentioned hooks in the template directory (e.g. for
new clones). You can still overwrite selectively for a repository by
[installing another set of hooks](#installing-or-removing-run-wrappers). Missing
Git LFS hooks will always be placed if necessary.

If you want, you can try out what the script would do first, without changing
anything by using:

```shell
$ cli installer --dry-run
```

### Install Mode: Centralized Hooks

Lastly, you have the option to install the templates to a centralized location
(`core.hooksPath`). You can read more about the difference between this option
and the default one [below](#templates-or-central-hooks). For this, run the
command below.

```shell
$ cli installer --use-core-hookspath
```

Optionally, you can also pass the template directory to which you want to
install the centralized hooks by appending `--template-dir <path>` to the
command above, for example:

```shell
$ cli installer --use-core-hookspath --template-dir /home/public/.githooks
```

### Install from different URL and Branch

If you want to install from another Git repository (e.g. from your own or your
companies fork), you can specify the repository clone url as well as the branch
name (default: `main`) when installing with:

```shell
$ cli installer --clone-url "https://server.com/my-githooks-fork.git" --clone-branch "release"
```

The installer always maintains a Githooks clone inside `<installDir>/release`
for its automatic update logic. The specified custom clone URL and branch will
then be used for further updates in the above example (see
[update machanics](#update-mechanics)).

Because the installer **always** downloads the latest release (here from another
URL/branch), it needs deploy settings to know where to get the binaries from.
Either your fork has setup these settings in their Githooks release (you
hopefully downloaded) already or you can specify them by using
`--deploy-api <type>` or the full settings file `--deploy-settings <file>`. The
`<type>` can either be `gitea` ( or `github` which is not needed since it can be
auto-detected from the URL) and it will automatically download and **verify**
the binaries over the implemented API. Credentials will be collected over
[`git credential`](https://git-scm.com/docs/cli/git-credential) to access the
API. [@todo].

### No Installation

You can use this hook manager also without a global installation. For that you
can clone this repository anywhere (e.g. `<repoPath>`) and build the executables
with Go by running `githooks/scripts/build.sh --prod`. You can then use the
hooks by setting `core.hooksPath` (in any suitable Git config) to the checked in
run-wrappers in `<repoPath>/hooks` like so:

```shell
git clone https://github.com/gabyx/githooks.git githooks
cd githooks
scripts/build.sh
git config --global core.hooksPath "$(pwd)/hooks"
```

### Non-Interactive Installation

You can also run the installation in **non-interactive** mode with the command
below. This will determine an appropriate template directory (detect and use the
existing one, or use the one passed by `--template-dir`, or use a default one),
install the hooks automatically into this directory, and enable periodic update
checks.

The global install prefix defaults to `${HOME}` but can be changed by using the
options `--prefix <installPrefix>`:

```shell
$ cli installer --non-interactive [--prefix <installPrefix>]
```

It's possible to specify which template directory should be used, by passing the
`--template-dir <dir>` parameter, where `<dir>` is the directory where you wish
the templates to be installed.

```shell
$ cli installer --template-dir "/home/public/.githooks-templates"
```

By default the script will install the hooks into the `~/.githooks/templates/`
directory.

### Install on the Server

On a server infrastructure where only _bare_ repositories are maintained, it is
best to maintain only server hooks. This can be achieved by installing with:

```shell
$ cli installer ---maintained-hooks "server"
```

The global template directory then **only** contains the following run-wrappers
for Githooks:

- `pre-push`
- `pre-receive`
- `update`
- `post-receive`
- `post-update`
- `reference-transaction`
- `push-to-checkout`
- `pre-auto-gc`

which get deployed with `git init` or `git clone` automatically. See also the
[setup for bare repositories](#setup-for-bare-repositories).

#### Setup for Bare Repositories

If you want to use Githooks with bare repositories on a server, you should setup
the following to ensure smooth operations (see [user prompts](#user-prompts)):

```shell
cd bareRepo
# Install Githooks into this bare repository
git hooks install

# Automatically accept changes to all existing and new
# hooks in the current repository.
# This makes the fatal trust prompt pass.
git hooks config trust-all-hooks --accept

# Don't do global automatic updates, since the Githooks updater
# might get invoked in parallel on a server.
git hooks config update --disable
```

Enabling auto-updates on the server, is a todo and needs file locks.

Note: A user cannot change bare repository Githooks by pushing changes to a bare
repository on the server. If you use shared hook repositories in you bare
repository, you might consider disabling shared hooks updates by
[`git hooks config disable-shared-hooks-update --set`](docs/cli/git_hooks_config_disable-shared-hooks-update).

### Templates or Global Hooks

This installer command can work in one of 2 ways:

- Using the git template folder `init.templateDir` (default behavior)
- Using the git `core.hooksPath` variable (set by passing the
  `--use-core-hookspath` parameter to the install script)

Read about the differences between these 2 approaches below.

In both cases, the installer command will make sure Git will find the Githooks
run-wrappers.

#### Template Folder (`init.templateDir`)

In this approach, the install script creates hook templates (global Git config
`init.templateDir`) that are installed into the `.git/hooks` folders
automatically on `git init` and `git clone`. For bare repositories, the hooks
are installed into the `./hooks` folder on `git init --bare`. This is the
recommended approach, especially if you want to selectively control which
repositories use Githooks or not.

The install script offers to search for repositories to which it will install
the run-wrappers, and any new repositories you clone will have these hooks
configured.

You can disable installing Githooks run-wrappers by using:

```shell
git clone --template= <url> <repoPath>
git lfs install # Important if you use Git LFS!. It never hurts doing this.
```

**Note**: It's recommended that you do `git lfs install` again. However, with
the latest `git` version 2.30, and `git lfs` version 2.9.2, `--template=` will
not result in **no** LFS hooks inside `${GIT_DIR}/hooks` if your repository
**contains** LFS objects.

#### Global Hooks Location (`core.hooksPath`)

In this approach, the install script installs the hook templates into a
centralized location (`~/.githooks/templates/` by default) and sets the global
`core.hooksPath` variable to that location. Git will then, for all relevant
actions, check the `core.hooksPath` location, instead of the default
`${GIT_DIR}/hooks` location.

This approach works more like a _blanket_ solution, where **all
repositories**<span id="a2">[<sup>2</sup>](#2)</span> will start using the hook
templates, regardless of their location.

**<span id="2"><sup>2</sup></span>[⏎](#a2) Note:** It is possible to override
the behavior for a specific repository, by setting a local `core.hooksPath`
variable with value `${GIT_DIR}/hooks`, which will revert Git back to its
default behavior for that specific repository. You don't need to initialize
`git lfs install`, because they presumably be already in `${GIT_DIR}/hooks` from
any `git clone/init`.

### Updates

You can update the Githooks any time by running one of the install commands
above. It will update itself and simply overwrite the template run-wrappers with
the new ones, and if you opt-in to install into existing or registered local
repositories, those will get overwritten too.

You can also enable automatic update checks during the installation, that is
executed **once a day after a successful commit**. It checks for a new version
and asks whether you want to install it. It then downloads the binaries (GPG
signed + checksummed) and dispatches to the new installer to install the new
version.

Automatic updates can be enabled or disabled at any time by running the command
below.

```shell
# enable with:
$ git hooks update --enable # `Config: githooks.autoUpdateEnabled`

# disable with:
$ git hooks update --disable
```

#### Update Mechanics

The update machanism works by tracking the tags on the Git branch (chosen at
install time) which is checked out in `<installDir>/release`. Normally, if there
are new tags (versions) available, the newest tag (version) is installed.
However, [prerelease version](https://semver.org) tags (e.g. `v1.0.0-rc1`) are
generally skipped. You can disable this behavior by setting the global Git
config value `githooks.autoUpdateUsePrerelease = true`. Major version updates
are **never** automatically installed an need the consent of the user.

If the annotated version tag or the commit message it points to
(`git tag -l --format=%(contents) <tag>`) contains a trailing header which
matches the regex `^Update-NoSkip: *true`, than this version **will not be
skipped**. This feature enables to enforce an update to a specific version. In
some cases this is useful (serialization changes etc.).

The single-line commit trailers `^Update-Info: *(.*)` on version tagged commits
are used to assemble a small changelog during update, which is presented to the
user. The single line can contain important information/links to relevant fixes
and changes.

You can also check for updates at any time by executing
[`git hooks update`](docs/cli/git_hooks_update.md) or using
[`git hooks config update [--enable|--disable]`](docs/cli/git_hooks_config_update.md)
command to enable or disable the automatic update checks.

## Uninstalling

If you want to get rid of this hook manager, you can execute the uninstaller
`<installDir>/bin/uninstaller` by

```shell
$ git hooks uninstaller
```

This will delete the run-wrappers installed in the template directory,
optionally the installed hooks from the existing local repositories, and
reinstates any previous hooks that were moved during the installation.

## YAML Specifications

You can find YAML examples for hook ignore files `.ignore.yaml` and shared hooks
config files `.shared.yaml` [here](docs/yaml-specs.md).

## Migration

Migrating from the `sh` [implementation here](https://github.com/gabyx/githooks)
is easy, but unfortunately we do not yet provide an migration option during
install (PRs welcome) to take over Git configuration values and other not so
important settings.

However, you can take the following steps for your old `.shared` and `.ignore`
files inside your repositories to make them work directly with a new install:

1. Convert all entries in `.ignore` files to a pattern in a YAML file
   `.ignore.yaml` (see [specs](#yaml-specifications)). Each old glob pattern
   needs to be prepended by `**/` (if not already existing) to make it work
   correctly (because of namespaces), e.g. a pattern `.*md` becomes `**/.*md`.
   Disable shared repositories in the old version need to be reconfigured, by
   using ignore patterns. Check if the ignore is working by running
   `[git hooks list](docs/cli/git_hooks_list.md)`.

2. Convert all entries in `.shared` files to an url in a YAML file
   `.shared.yaml` [here](docs/yaml-spec.md).

3. It's heartly recommended to **first** uninstall the old version, to get rid
   of any old settings.

4. Install the new version.

Trusted hooks will be needed to be trusted again. To port Git configuration
variables use the file `githooks/hooks/gitconfig.go` which contains all used Git
config keys.

## Dialog Tool

Githooks provides it's own **platform-independent dialog tool `dialog`** which
is located in `<installDir>/bin`. It enables the use of **native** GUI dialogs
such as:

- options dialog
- entry dialog
- file save and file selection dialogs
- message dialogs
- system notifications

inside of hooks and scripts.
**[See the screenshots.](docs/dialog-screenshots/Readme.md)**

_Why another tool?:_ At the moment of writting there exists no proper
platform-independent GUI dialog tool which is **bomb-proof in it's output and
exit code behavior**. This tool should really enable proper and safe usage
inside hooks and other scripts. You can even report the output in `json` format
(use option `--json`). It was heavily inspired by
[zenity](https://github.com/ncruces/zenity) and features some of the same
properties (no `cgo`, cancelation through `context`). You can use this dialog
tool independently of Githooks.

**Test it out!** 🎉: Please refer to the
[documentation of the tool](docs/dialog/dialog.md).

### Build From Source

```shell
cd githooks
go mod download
go mod vendor
cd githooks/apps/dialog
go build ./...
./dialog --help
```

### Dependencies

The dialog tool has the following dependencies:

- `macOS` : `osascript` which is provided by the system directly.
- `Unix` : A dialog tool such as `zenity` (preferred), `qarma` or `matedialog`.
- `Windows`: Common Controls 6 which is provided by the system directly.

## Tests and Debugging

Running the integration tests with Docker:

```shell
cd githooks
bash tests/test-alpine.sh # and other 'test-XXXX.sh' files...
```

Run certain tests only:

```shell
bash tests/test-alpine.sh --seq {001..120}
bash tests/test-alpine.sh --seq 065
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

### Todos:

- Finish deploy settings implementation for Gitea and others.

## Changelog

### Version v2.x.x

For upgrading from `v1.x.x` to `v2.x.x` consider the
[braking change documention](docs/changes/Braking-Changes-v2.md).

## FAQ

- **Shell on Windows shows weird characters:** Githooks outputs UTF-8 characters
  (emojis etc.). Make sure you have the UTF-8 codepage active by doing
  `chcp.com 65001` (either in `cmd.exe` or `git-bash.exe`, also from an
  integrated terminal in VS Code). You can make it permanent by putting this
  into the startup scripts of your shell, e.g. (`.bashrc`). Consider using
  Windows Terminal.

## Acknowledgements

- [Original Githooks implementation in `sh`](http://github.com/rycus86/githooks)
  by Viktor Adam.

## Authors

- Gabriel Nützi (`Go` implementation)
- Viktor Adam (Initial `sh` implementation)
- Matthijs Kooijman (suggestions & discussions)
- and community.

## Support & Donation

When you use Githooks and you would like to say thank you for its development
and its future maintenance: I am happy to receive any donation which will be
distributed equally among all contributors.

[![paypal](https://www.paypalobjects.com/en_US/i/btn/btn_donate_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=6S6BJL4GSMSG4)

## License

MIT
