# API Breaking Changes from Version `2` to `3`

**Note: You cannot update to version `3` from version `2`. Uninstall version `2`
with `git hooks uninstaller` first.**

- Githooks does not **run** updates **automatically** anymore.
- Update checks are however still executed if `git hooks config update-checks`
  is enabled. The following settings are changed:

- **Only two install modes are supported:**:
  - The **normal mode** sets **local** `core.hooksPath` on `git hooks install`
    in a repository and will not do any automatic installation on clone/init.

  - The **centralized mode** sets this globally and will work by default for all
    clone/init repos.

- Run-wrappers can still be installed in local repos, **`init.templateDir` is
  not used and controlled by Githooks anymore**.

- Run-wrapper will use by default `githooks-runner` and as fallback use
  `git config githooks.runner`. Package manager builds
  `tag: package_manager_enabled` will not set this.

- The installer has an option `--hooks-dir` which specifies the directory to
  place the maintained hooks.

- The installer has an option `--hooks-dir-use-template-dir` in
  non-`--centralized` mode which looks for a set `GIT_TEMPLATE_DIR` or
  `init.templateDir` or the Git default directory. This option is not
  encouraged, since installing hooks in such a directory use by Git will install
  templates (and therefore the hooks) on each init/clone as before.

- The installer warns if the chosen hooks directory during install is pointing
  to a template directory used by Git.

- User wanting to install run-wrappers inside a repository instead of setting
  `core.hooksPath` should either use `git hooks install --maintained-hooks ...`
  or place a file `<template-dir>/hooks/githooks-contains-run-wrappers` to let
  Githooks know that this repo maintains run-wrappers (for updates etc, that no
  `core.hooksPath` is used but run-wrappers are installed directly).

- Registering is now done for all repositories (also in centralized mode).

- `git hook uninstall --full` and
  `git hooks uninstaller --full-uninstall-from-repos` will clean all Git config
  and cached settings (checksums) in registered repositories. By default
  `git hooks uninstall` will not remove locally set `githooks.*` Git config
  variables, this is to make `reinstallation` more easy, e.g.
  `githooks.maintainedHooks` stays and will be read on reinstall.
