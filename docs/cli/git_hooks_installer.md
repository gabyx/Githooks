## git hooks installer

Githooks installer application.

### Synopsis

Githooks installer application. It downloads the Githooks artifacts of the
current version from a deploy source and verifies its checksums and signature.
Then it calls the installer on the new version which will then run the
installation procedure for Githooks.

See further information at https://github.com/gabyx/githooks/blob/main/README.md

```
git hooks installer [flags]
```

### Options

```
      --log string                              Log file path (only for installer).
      --dry-run                                 Dry run the installation showing what's being done.
      --non-interactive                         Run the installation non-interactively
                                                without showing prompts.
      --update                                  Install and update directly to the latest
                                                possible tag on the clone branch.
      --skip-install-into-existing              Skip installation into existing repositories
                                                defined by a search path.
      --prefix string                           Githooks installation prefix such that
                                                `<prefix>/.githooks` will be the installation directory.
      --hooks-dir ~/.githooks/templates/hooks   The preferred directory to use for maintaining Githook's hook run-wrappers
                                                e.g. ~/.githooks/templates/hooks.
      --hooks-dir-use-template-dir              If the `GIT_TEMPLATE_DIR` env. variable or `init.templateDir`
                                                or the Git default template directory is used in place of `--hooks-dir`.
      --maintained-hooks strings                A set of hook names which are maintained in the template directory.
                                                Any argument can be a hook name `<hookName>`, `all` or `server`.
                                                An optional prefix '!' means subtraction from the current set.
                                                The initial value of the internally built set defaults
                                                to all hook names if `all` or `server` (or negated) is not given
                                                as first argument:
                                                  - `all` : All hooks supported by Githooks.
                                                  - `server` : Only server hooks supported by Githooks.
                                                You can list them separately or comma-separated in one argument.
      --centralized                             If the install mode `centralized` should be used which
                                                sets the global `core.hooksPath`.
      --git-config-no-abs-path                  Make certain Githooks Git config values
                                                not use abs. paths. Useful to have the Git config not change.
                                                This means you need to have the Githooks binaries in your path.
      --clone-url string                        The clone url from which Githooks should clone
                                                and install/update itself. Githooks tries to
                                                auto-detect the deploy setting for downloading binaries.
                                                You can however provide a deploy settings file yourself if
                                                the auto-detection does not work (see `--deploy-settings`).
      --clone-branch string                     The clone branch from which Githooks should
                                                clone and install/update itself.
      --deploy-api string                       The deploy api type (e.g. [`gitea`, `github`]) to use for updates
                                                of the specified `clone-url` for helping the deploy settings
                                                auto-detection. For Github urls, this is not needed.
      --deploy-settings string                  The deploy settings YAML file to use for updates of the specified
                                                `--clone-url`. See the documentation for further details.
      --build-from-source                       If the binaries are built from source on updates instead of
                                                downloaded from the deploy url.
      --build-tags strings                      Build tags for building from source (get extended with defaults).
                                                You can list them separately or comma-separated in one argument.
      --use-pre-release                         When fetching the latest installer, also consider pre-release versions.
  -h, --help                                    help for installer
```

### SEE ALSO

- [git hooks](git_hooks.md) - Githooks CLI application

###### Auto generated by spf13/cobra
