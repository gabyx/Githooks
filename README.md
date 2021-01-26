<img src="docs/githooks-logo.svg" style="margin-right: 20pt" align="left">
<h1>Githooks <span style="font-size:12pt">on Steroids</span></h1>

[![Build Status](https://travis-ci.org/gabyx/githooks.svg?branch=main)](https://travis-ci.org/gabyx/githooks)
[![Coverage Status](https://coveralls.io/repos/github/gabyx/githooks/badge.svg?branch=main)](https://coveralls.io/github/gabyx/githooks?branch=main)

**STILL BETA: Any changes with out notice!**
<br><br><br><br>

A **platform-independend hooks manager** written in Go to support shared hook repositories and per-repository [Git hooks](https://git-scm.com/docs/cli/githooks), checked into the working repository. This implementation is the Go port and successor of the [original impementation](https://github.com/rycus86/githooks) (see [Migration](#migrating)).

To make this work, the installer creates run-wrappers for Githooks that are installed into the `.git/hooks`
folders automatically on `git init` and `git clone`. There's more [to the story though](#templates-or-global-hooks).
When one of the Githooks run-wrappers executes, Githooks starts up and tries to find matching hooks in the
`.githooks` directory under the project root, and invoke them one-by-one.
Also it searches for hooks in configured shared hook repositories.

**This Git hook manager supports:**

- Running repository checked-in hooks.
- Running shared hooks from other Git repositories (with auto-update).
- Command line interface.
- Fast execution due to compiled Go executable.
- Fast parallel execution over threadpool (not yet finished).
- Ignoring non-shared and shared hooks with patterns.
- Automatic Githooks updates:
  Fully configurable for your own company by url/branch and deploy settings.

## Layout and Options

Take this snippet of a project layout as an example:

```
/
└── .githooks/
    └── commit-msg/
        ├── validate
        └── add-text
    └── pre-commit/
        ├── 01-validate
        ├── 02-lint
        ├── 03-test.yaml
        ├── docs.md
        └── .ignore.yaml
    └── post-checkout
    └── ...
    └── .ignore.yaml
    └── .shared.yaml
    └── .lfs-required
├── README.md
└── ...
```

All hooks to be executed live under the `.githooks` top-level folder, that should be checked into the repository.
Inside, we can have directories with the name of the hook (like `commit-msg` and `pre-commit` above),
or a file matching the hook name (like `post-checkout` in the example). The filenames in the directory
do not matter, but the ones starting with a `.` (dotfiles) will be excluded by default.
All others are executed in lexical order alphabetical order
according to the Go function [`Walk`](https://golang.org/pkg/path/filepath/#Walk).
rules.

You can use the [command line helper](docs/cli/git_hooks.md) tool as `git hooks list`
(a globally configured Git alias `alias.hooks`) to list all the hooks that apply to
the current repository and their current state.

## Execution

If a file is executable, it is directly invoked, otherwise it is interpreted with the `sh` shell.
On Windows that mostly means dispatching to the `bash.exe` from [https://gitforwindows.org](https://gitforwindows.org).

**All parameters and standard input** are forwarded from Git to the hooks.
The standard output and standard error of any hook which Githooks runs is captured **together**<span id="a1">[<sup>1</sup>](#1)</span>. and printed to the standard error stream which might or might not get read by Git itself (e.g. `pre-push`).

Hooks can also be specified by a run configuration in a corresponding YAML file,
see [#hook-run-configuration](Hook Run Configuration).

Hooks related to `commit` events (where it makes sense, not `post-commit`) will also have a `${STAGED_FILES}`
 environment variable setthat is the list of staged and changed files according to
 `git diff --cached --diff-filter=ACMR --name-only`.  File paths are separated by a newline `\n`.
If you want to iterate in a shell script over them,
and expect spaces in paths, you might want to set the `IFS` like this:

```shell
IFS="
"
for STAGED in ${STAGED_FILES}; do
    ...
done
```

The `ACMR` filter in the `git diff` will include staged
files that are added, copied, modified or renamed.

**<span id="1"><sup>1</sup></span>[⏎](#a1) Note:** This caveat is basically there because stdandard output and error might get interleaved badly and so far no solution to this small problem has been tackled yet. It is far better to output both streams in the correct order, and therefore send it the error stream because that will not conflict in anyway with Git (see [fsmonitor-watchman](https://git-scm.com/docs/githooks#_fsmonitor_watchman), unsupported right now.). If that poses a real problem for you, open an issue.

### Hook Run Configuration

Each supported hook can also be specified by a configuration file `<hookName>.yaml` where `<hookName>` is any [supported hook name](#supported-hooks). An example might look like the following:

```yaml
# The command to run.
cmd: "my-command.exe"

# The arguments given to `cmd`.
args:
  - "-s"
  - "--all"

# If you want to make sure your file is not
# treated always as the newest version. Fix the version by:
version: 1
```

All additional arguments given by Git to `<hookName>` will be appended last onto `args`.

**Sidenote**: You might ask why we split this configuration into one for each hook instead of one collocated YAML file. The reason is that each hook invocation by Git is separate. Avoiding reading this total file several times needs time and since we want speed and only an opt-in solution this is avoided.

## Supported Hooks

The supported hooks are listed below.
Refer to the [Git documentation](https://git-scm.com/docs/cli/githooks)
for information on what they do and what parameters they receive.

- `applypatch-msg`
- `pre-applypatch`
- `post-applypatch`
- `pre-commit`
- `pre-merge-commit`
- `prepare-commit-msg`
- `commit-msg`
- `post-commit`
- `pre-rebase`
- `post-checkout`
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

The `fsmonitor-watchman` hook is currently not supported.
If you have a use-case for it and want to use it with this tool, please open an issue.

## Git Large File Storage (Git LFS) Support

If the user has installed [Git Large File Storage](https://git-lfs.github.com/) (`git-lfs`) by calling
`git lfs install` globally or locally for a repository only,
`git-lfs` installs 4 hooks when initializing (`git init`) or cloning (`git clone`) a repository:

- `post-checkout`
- `post-commit`
- `post-merge`
- `pre-push`

Since Githooks overwrites the hooks in `.git/hooks`, it will also run all *Git LFS* hooks internally
if the `git-lfs` executable is found on the system path. You can enforce having `git-lfs` installed on
the system by placing a `./githooks/.lfs-required` file inside the repository, then if `git-lfs` is missing,
a warning is shown and the hook will exit with code `1`. For some `post-*` hooks this does not mean that the
outcome of the git command can be influenced even tough the exit code is `1`, for example `post-commit` hooks
can't fail commits. A clone of a repository containing this file might still work but would issue a warning
and exit with code `1`, a push - however - will fail if `git-lfs` is missing.

It is advisable for repositories using *Git LFS* to also have a pre-commit
hook (e.g. `examples/lfs/pre-commit`) checked in which enforces a correct installation of _Git LFS_.

## Shared hook repositories

The hooks are primarily designed to execute programs or scripts in the `.githooks` folder of a single repository.
However there are use-cases for common hooks, shared between many repositories with similar requirements and functionality.
For example, you could make sure Python dependencies are updated on projects that have a `requirements.txt` file,
or an `mvn verify` is executed on `pre-commit` for Maven projects, etc.

For this reason, you can place a `.shared.yaml` file (see [specs](#yaml-specification)) inside the `.githooks` repository, which can hold a list of repositories which hold common and shared hooks. Alternatively, you can have a shared repositories set by multiple `githooks.shared` local or global Git configuration variables, and the hooks in these repositories will execute for all local projects where Githooks is installed.
See [git hooks shared](docs/cli/git_hooks_shared.md) for configuring all 3 types of shared hooks repositories.

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

A example config `<repoPath>/.githooks/shared.yaml` (see [specs](#yaml-specification)):

```yaml
version: 1
urls:
  - ssh://user@github.com/shared/special-hooks.git@otherbranch
  - git@github.com:shared/repo.git@mybranch
```

The install script offers to set up shared hooks in the global Git config.
but you can do it any time by changing the global configuration variable.

### Supported URLS

Supported URL for shared hooks are:

- **All URLs [Git supports](https://git-scm.com/docs/cli/git-clone#_git_urls)** such as:

  - `ssh://github.com/shared/hooks-maven.git@mybranch` and also the short `scp` form
    `git@github.com:shared/hooks-maven.git`
  - `git://github.com/shared/hooks-python.git`
  - `file:///local/path/to/bare-repo.git@mybranch`

  All URLs can include a tag specification syntax at the end like `...@<tag>`, where `<tag>` is a Git tag, branch or commit hash.
  The `file://` protocol is treated the same as a local path to a bare repository, see *local paths* below.

- **Local paths** to bare and non-bare repositories such as:

  - `/local/path/to/checkout` (gets used directly)
  - `/local/path/to/bare-repo.git` (gets cloned internally)

  Note that relative paths are relative to the path of the repository executing the hook.
  These entries are forbidden for **shared hooks** configured by `.githooks/.shared.yaml` per repository
  because it makes little sense and is a security risk.

Shared hooks repositories specified by *URLs* and *local paths to bare repository* will be checked out into the `<instalPrefix>/.githooks/shared` folder
(`~/.githooks/shared` by default), and are updated automatically after a `post-merge` event (typically a `git pull`)
on any local repositories. Any other local path will be used **directly and will not be updated or modified**.
Additionally, the update can also be triggered on other hook names by setting a comma-separated list of additional
hook names in the Git configuration parameter `githooks.sharedHooksUpdateTriggers` on any configuration level.

An additional global configuration parameter `githooks.failOnNonExistingSharedHooks` makes hooks fail with an error if any shared hook configured in `.shared.yaml` is missing, meaning `git hooks update` has not yet been called.
See [`git hooks config fail-on-non-existing-shared-hooks --help`](docs/cli/git_hooks_config_fail-on-non-existing-shared-hooks.md)

You can also manage and update shared hook repositories using the [`git hooks shared --help`](docs/cli/git_hooks_shared.md) tool.

## Layout of Shared Hook Repositories

The layout of these shared repositories is the same as above, with the exception that the hook folders (or files) can be at the project root as well, to avoid the redundant `.githooks` folder.

If you want the shared hook repository to use Githooks itself (e.g. for development purposes by using hooks from `<sharedRepo>/.githooks`) you can furthermore place the *shared* hooks inside a `<sharedRepo/githooks` subfolder.
In that case the `<sharedRepo>/.githooks` folder is ignored when other users use this shared repository.

So the priority to find hooks in a shared hook repository is as follows: consider hooks

1. in `<sharedRepo>/githooks`, if it does not exist, consider hooks in
2. in `<sharedRepo>/.githooks`, if it does not exist consider hooks
3. in `<sharedRepo>` as the last fallback.

Each of these directories can be of the same format as the normal `.githooks` folder in a single repository.

### Shared Repository Namespace

A shared repository can optionally have a namespace associated with it. The name can be stored in a file `.namespace` in the hooks directory of the shared repository, e.g. one of the following:

- `<sharedRepo>/githooks/.namespace` if the shared hooks are inside `<sharedRepo>/githooks`
  (because you use the local hooks `.githooks` for the development of this shared repository).
- `<sharedRepo>/.githooks/.namespace` if the shared hooks are inside `<sharedRepo>/.githooks`.
- `<sharedRepo>/.namespace` if the shared hooks are at the root of the repository.

## Ignoring Hooks and Files

The `.ignore.yaml` (see [specs](#yaml-specification)) files allow excluding files

- from being treated as hook scripts or
- hooks from beeing run.

They allow *glob* filename patterns (with double-star `**` syntax to match multiple directories) and
paths to be matched against a hook's (file's) *namespace path* which consists of
the name of the hook prefixed by a hook namespace , e.g. `<hookNamespace>/<hookName>`.
A [namespace](#shared-repository-namespace) comes into play when the hook (or file) belongs to a shared hook repository.]

You can ignore executing all sorts of hooks per Git repository by specifying patterns
or paths to ignore which match against this namespace path.

Each hook either in the current repository `.githooks/...` or inside a shared hooks
repository has a so called *namespace path*. All ignore entries (patterns or paths) will match against these
paths. Each shared repository can provide a namespace (see )
You can inspect all *namespace paths* by inspecting `ns-path:` in the output of [git hooks list](docs/cli/git_hooks_list.md) in the current repository.

```shell
# Disable certain hooks by a pattern in this repository:
# User ignore pattern stored in `.git/.githooks.ignore.yaml`:
$ git hooks ignore add --pattern "pre-commit/**" # Store: `.git/.githooks.ignore.yaml`:

# Disable certain shared hooks (with namespace 'my-shared-super-hooks')
# by a pattern in this repository:
$ git hooks ignore add --pattern "my-shared-super-hooks/pre-commit/**"
```

In the above example, one of the `.ignore.yaml` files should contain a pattern `**/*.md` to exclude the `pre-commit/docs.md` Markdown file.

The `.githooks/.ignore.yaml` file applies to each of the hook directories, and should still define filename patterns, `*.txt` instead of `**/*.txt` for example. If there is a `.ignore.yaml` file both in the hook type folder and in `.githooks`, the files whose filename matches any pattern from either of those two files will be excluded. You can also manage `.ignore.yaml` files using [`git hooks ignore [add|remove] --help`](docs/cli/git_hooks_ignore.md), and consult this help for futher information on pattern syntax.

Hooks in individual shared repositories can be disabled as well, running [`git hooks ignore [add|remove] --help`](docs/cli/git_hooks_ignore_add.md)` by specifing patterns or namespace paths.

Finally, all hook execution can be bypassed with a non-empty value in the `$GITHOOKS_DISABLE` environment variable too.

## Trusting Hooks

To try and make things a little bit more secure, Githooks checks if any new hooks were added we haven't run before,
or if any of the existing ones have changed. When they have, it will prompt for confirmation whether you accept
those changes or not, and you can also disable specific hooks to skip running them until you decide otherwise.
The accepted checksums are maintained in the `.git/.githooks.checksum` directory, per local repository.
You can however use a global checksum directory setup by specifing `githooks.checksumCacheDir`
in any suitable Git config (can be different for each repository).

If the repository contains a `.githooks/trust-all` file, it is marked as a trusted repository. On the first interaction with hooks, Githooks will ask for confirmation that the user trusts all existing and future hooks in the repository,
and if she does, no more confirmation prompts will be shown.
This can be reverted by running either the
`git config --unset githooks.trustAll`, or the [`git hooks config trusted --help`](docs/cli/git_hooks_config_trusted.md) command.
This is a per-repository setting.
Consult [`git hooks trust --help`](docs/cli/git_hooks_trust.md) and [`git hooks config trusted --help`](docs/cli/git_hooks_config_trusted.md)
for more information.

There is a caveat worth mentioning: if a terminal *(tty)* can't be allocated, then the default action is to accept the changes or new hooks.
A terminal cannot be allocated for exmaple if you execute Git over a GUI such as VS Code or any other Git GUI.
Let me know in an issue if you strongly disagree, and you think this is a big enough risk worth having slightly worse UX instead.

You can also accept (trust) hooks by using [`git hooks trust hooks ---help`](docs/cli/git_hooks_trust_hooks.md).

## Disabling Githooks

To disable running any Githooks locally or globally, use the following:

```shell
# Disable Githooks completely for this repository:
$ git hooks disable # Use --reset do undo.
# or
$ git hooks config disable --set # Same thing... Config: `githooks.disable`


# Disable Githooks globally:
$ git hooks disable --global # Use --reset do undo.
# or
$ git hooks config disable --set --global # Same thing... Config: `githooks.disable`
```

Also, as mentioned above, all hook executions can be bypassed with a non-empty value in the `$GITHOOKS_DISABLE` environment variable.

> See the documentation of the command line helper tool on its [docs page](https://github.com/rycus86/githooks/blob/master/docs/cli/command-line-tool.md)!

### Removing Run-Wrappers

You can install and uninstall run-wrappers inside a repository with [`git hooks install`](docs/cli/git_hooks_install.md).
or [`git hooks uninstall`](docs/cli/git_hooks_install.md).
This installs and uninstalls wrappers from `${GIT_DIR}/hooks` as well as sets and unsets local Githooks-internal Git configuration variables.

## Installation

- [Download the latest release](https://github.com/gabyx/githooks/releases), exctract it and execute the installer by the below instructions.

The installer will:

1. Find out where the Git templates directory is.
   1. From the `$GIT_TEMPLATE_DIR` environment variable.
   2. With the `git config --get init.templateDir` command.
   3. Checking the default `/usr/share/git-core/templates` folder.
   4. Search on the filesystem for matching directories.
   5. Offer to set up a new one, and make it `init.templateDir`.
2. Write all Githooks run-wrappers into the choosen directory:
   - either `init.templateDir` or
   - `core.hooksPath` depending on the install mode `--use-core-hooks-path`.
3. Offer to enable automatic update checks.
4. Offer to find existing Git repositories on the filesystem (disable with `--skip-install-into-existing`)
   1. Install run-wrappers into them (`.git/hooks`).
   2. Offer to add an intro README in their `.githooks` folder.
5. Install/update run-wrappers into all registered repositories:
   Repositories using Githooks get registered in the install folders `registered.yaml` file
   on their first hook invocation.
6. Offer to set up shared hook repositories.

### Normal Installation

To install Githooks on your system, simply execute the executable `installer`.
It will guide you through the installation process.
Check the `installer --help` for available options. Some of them are described below:

If you want, you can try out what the script would do first, without changing anything by using:

```shell
$ installer --dry-run
```

### Non-Interactive Installation

You can also run the installation in **non-interactive** mode with the command below. This will determine an appropriate template directory (detect and use the existing one, or use the one passed by `--template-dir`, or use a default one), install the hooks automatically into this directory, and enable periodic update checks.

The global install prefix defaults to `${HOME}` but can be changed by using the options `--prefix <installPrefix>`:

```shell
$ installer --non-interactive [--prefix <installPrefix>]
```

It's possible to specify which template directory should be used, by passing the `--template-dir <dir>` parameter, where `<dir>` is the directory where you wish the templates to be installed.

```shell
$ installer --template-dir "/home/public/.githooks-templates"
```

By default the script will install the hooks into the `~/.githooks/templates/` directory.

### Install Mode: Centralized Hooks
Lastly, you have the option to install the templates to, and use them from a centralized location. You can read more about the difference between this option and default one [below](#templates-or-central-hooks). For this, run the command below.

```shell
$ installer --use-core-hookspath
```

Optionally, you can also pass the template directory to which you want to install the centralized hooks by appending `--template-dir <path>` to the command above, for example:

```shell
$ installer --use-core-hookspath --template-dir /home/public/.githooks
```

### Install from different URL and Branch
If you want to install from another Git repository (e.g. from your own or your companies fork),
you can specify the repository clone url as well as the branch name (default: `main`) when installing with:

```shell
$ installer --clone-url "https://server.com/my-githooks-fork.git" --clone-branch "release"
```

Because the installer **always** downloads the latest release (here from another URL/branch), it needs deploy settings
to know where to get the binaries from. Either your fork has setup
these settings in their Githooks release (you hopefully downloaded) already or you can specify them by using
`--deploy-api <type>` or the full settings file `--deploy-settings <file>`.
The `<type>` can either be `gitea` ( or `github` which is not needed since it can be auto-detected from the URL) and
it will automatically download and **verify** the binaries over the implemented API.
Credentials will be collected over [`git credential`](https://git-scm.com/docs/cli/git-credential) to access the API. [@todo].

The clone URL and branch will then be used for further updates.

### Install on the Server

On a server infrastructure where only *bare* repositories are maintained, it is best to maintain only server hooks.
This can be achieved by installing with the additional flag `--only-server-hooks` by:

```shell
$ installer --only-server-hooks
```

The global template directory then **only** maintain contains the following run-wrappers for Githooks:

- `pre-push`
- `pre-receive`
- `update`
- `post-receive`
- `post-update`
- `reference-transaction`
- `push-to-checkout`
- `pre-auto-gc`

which get deployed with `git init` or `git clone` automatically.
See also the [setup for bare repositories](#setup-bare).

### Setup for Bare Repositories

Because bare repositories mostly live on a server, you should setup the following if
you use a shared hooks repository (can live on the same server, see [shared URLs](#supported-urls))

```shell
cd bareRepo
# Install Githooks into this bare repository
# which will only install server hooks:
git hooks install

# Creates `.githooks/trust-all` marker for this bare repo
# This is necessary to circumvent trust prompts for shared hooks.
git hooks trust

# Automatically accept changes to all existing and new
# hooks in the current repository.
git hooks config trusted --accept

# Don't do global automatic updates, since the Githooks update
# script should not be run in parallel on a server.
git hooks config update --disable
```

### Templates or Global Hooks

This installer can work in one of 2 ways:

- Using the git template folder `init.templateDir` (default behavior)
- Using the git `core.hooksPath` variable (set by passing the `--use-core-hookspath` parameter to the install script)

Read about the differences between these 2 approaches below.

In both cases, the installer will make sure Git will find the Githooks run-wrappers.

#### Template Folder (`init.templateDir`)

In this approach, the install script creates hook templates (global Git config `init.templateDir`) that are installed into the `.git/hooks` folders automatically on `git init` and `git clone`. For bare repositories, the hooks are installed into the `./hooks` folder on `git init --bare`.
This is the recommended approach, especially if you want to selectively control which repositories use Githooks or not.

The install script offers to search for repositories to which it will install the run-wrappers, and any new repositories you clone will have these hooks configured.

You can disable installing Githooks run-wrappers by using:

```shell
git clone --template= <url> <repoPath>
git lfs install # Important if you use Git LFS!. It never hurts doing this.
```

**Note**: It's recommended that you do `git lfs install` again. With the latest `git` version 2.30, and `git lfs` version 2.9.2, `--template=` will not result in **no** LFS hooks inside `${GIT_DIR}/hooks` if your repository **contains** LFS objects.

#### Global Hooks Location (`core.hooksPath`)

In this approach, the install script installs the hook templates into a centralized location (`~/.githooks/templates/` by default) and sets the global `core.hooksPath` variable to that location. Git will then, for all relevant actions, check the `core.hooksPath` location, instead of the default `${GIT_DIR}/hooks` location.

This approach works more like a *blanket* solution, where **all repositories**<span id="a2">[<sup>2</sup>](#2)</span> will start using the hook templates, regardless of their location.

**<span id="2"><sup>2</sup></span>[⏎](#a2) Note:** It is possible to override the behavior for a specific repository, by setting a local `core.hooksPath` variable with value `${GIT_DIR}/hooks`, which will revert Git back to its default behavior for that specific repository.
You don't need to initialize `git lfs install`, because they presumably be already in `${GIT_DIR}/hooks` from any `git clone/init`.

### Supported Platforms:

The following platforms are tested:

- Linux
- MacOs
- Windows

### Updates

You can update the scripts any time by running one of the install commands above. It will simply overwrite the run-wrappers with the new ones, and if you opt-in to install into existing local repositories, those will get overwritten too.

You can also enable automatic update checks during the installation, that is executed once a day after a successful commit. It checks for a new version and asks whether you want to install it. It then downloads the binaries and dispatches to the new installer to install the new version.

Automatic updates can be enabled or disabled at any time by running the command below.

```shell
# enable with:
$ git hooks update --enable # `Config: githooks.autoUpdateEnabled`

# disable with:
$ git hooks update --disable
```

You can also check for updates at any time by executing
[`git hooks update`](docs/cli/git_hooks_update.md) or using
[`git hooks config update [--enable|--disable]`](docs/cli/git_hooks_config_update.md) command to enable or disable the automatic update checks.

### Custom User Prompt

If you want to use a GUI dialog when Githooks asks for user input, you can use an executable or script file to display it.
The example in the `examples/tools/dialog` folder contains a Python script `run` which uses the Python provided `tkinter` to show a dialog.

```shell
# install the example dialog tool from this repository
$ git hooks tools register dialog "./examples/tools/dialog"
```

This will copy the tool to the Githooks install folderTO execute when displaying user prompts.
The tool's interface is as follows.

```shell
$ run <title> <text> <options> <long-options>    # if `run` is executable
$ sh run <title> <text> <options> <long-options> # otherwise, assuming `run` is a shell script
```

The arguments for the dialog tool are:

- `<title>` the title for the GUI dialog
- `<text>` the text for the GUI dialog
- `<short-options>` the button return values, separated by slashes, e.g. `Y/n/d`. The default button is the first capital character found.
- `<long-options>` the button texts in the GUI, e.g. `Yes/no/disable`

The script needs to return one of the short-options on the standard output.
If the exit code is not `0`, the normal prompt on the standard input is shown as a fallback mechanism.

**Note:** Githooks will probably in the future provide a default cross-platform Dialog implementation, which will render this feature obsolete. (PRs welcome, see [https://github.com/gen2brain/dlgs](https://github.com/gen2brain/dlgs))

## Uninstalling

If you want to get rid of this hook manager, you can execute the uninstaller `<installDir>/bin/uninstaller` by

```shell
$ git hooks uninstall --global
```

This will delete the run-wrappers installed in the template directory, optionally the installed hooks from the existing local repositories, and reinstates any previous hooks that were moved during the installation.

## YAML Specifications

You can find YAML examples for hook ignore files `.ignore.yaml` and shared hooks config files `.shared.yaml` [here](docs/yaml-spec.md).

## Migration

Migrating from the `sh` [implementation here](https://github.com/rycus86/githooks) is easy, but unfortunately we do not yet provide an migration option during install (PRs welcome) to take over Git configuration values
and other not so important settings.

However, you can take the following steps for your old `.shared` and `.ignore` files inside your repositories to make them work
directly with a new install:

1 Convert all entries in `.ignore` files to a pattern in a YAML file `.ignore.yaml` (see [specs](#yaml-specification)). Each old glob pattern needs to be prepended by
  `**/` (if not already existin) to make it work correctly (because of namespaces), e.g. a pattner `.*md` becomes `**/.*md`.
  Disable shared repositories in the old version need to be reconfigured, be ignore patterns.
  Check if the ignore is working by running `[git hooks list](docs/cli/git_hooks_list.md)`.

2. Convert all entries in `.shared` files to an url in a YAML file `.shared.yaml`.

3. It's heartly recommended to **first** uninstalling the old version, to get rid of any old settings.

4. Install the new version.

Trusted hooks will be needed to be trusted again.
To port Git configuration variables use the file `githooks/hooks/gitconfig.go` which contains all used Git config keys.

## Acknowledgements

- [Original Githooks implementation in `sh`](http://github.com/rycus86/githooks) by Victor Adams.

## Authors

- Gabriel Nützi (`go` implementation)
- Viktor Adams (Initial `sh` implementation)
- and community.

## License

MIT