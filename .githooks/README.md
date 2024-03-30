# Githooks

This project uses [Githooks](https://github.com/gabyx/githooks), that allows
running [Git hooks](https://git-scm.com/docs/cli/githooks) checked into this
repository. This folder contains hooks that should be executed by everyone who
interacts with this source repository. For a documentation on how this works and
how to get it [installed](https://github.com/gabyx/githooks#installation), check
the project [README](https://github.com/gabyx/githooks/blob/main/README.md) in
the [gabyx/githooks](https://github.com/gabyx/githooks) GitHub repository.

## Brief summary

The [directories or files](https://github.com/gabyx/githooks#layout-and-options)
in this folder tell Git to execute certain scripts on various
[trigger events](https://github.com/gabyx/githooks#supported-hooks), before or
after a commit, on every checkout, before a push for example - assuming
[Githooks](https://github.com/gabyx/githooks) is already
[installed](https://github.com/gabyx/githooks#installation) and
[run-wrappers are installed](https://github.com/gabyx/githooks#installing-or-removing-run-wrappers)
for the repository. The directory or file names refer to these events, like
`pre-commit`, `post-commit`, `post-checkout`, `pre-push`, etc. If they are
folders, each file inside them is treated as a hook script and will be executed
when Git runs the hooks as part of the command issued by the user.
[Githooks](https://github.com/gabyx/githooks) comes with a
[command line tool](https://github.com/gabyx/githooks/blob/main/docs/cli/git_hooks.md)
tool, that allows you to manage its configuration and state with a
`git hooks <cmd>` command. See the
[documentation](https://github.com/gabyx/githooks/blob/main/docs/cli/git_hooks.md)
or run `git hooks --help` for more information and available options.

### Is this safe?

[Githooks](https://github.com/gabyx/githooks) uses an
[opt-in model](https://github.com/gabyx/githooks#trusting-hooks), where it will
ask for confirmation whether new or changed hooks should be run or not (or
disabled).

### How do I add a new hook script?

Either create a file with the
[Git hook](https://github.com/gabyx/githooks#supported-hooks) name, or a
directory (recommended) inside the `.githooks` folder, and place files with the
individual steps that should be executed for that event inside. If the file is
executable, it will be invoked directly, otherwise it is assumed to be a Shell
script - unless this file matches one of the
[ignore patterns](https://github.com/gabyx/githooks#ignoring-hooks-and-files) in
the `.githooks` area.

### How can I see what hooks are active?

You can look at the `.githooks` folder to see the local hooks in the repository,
though if you have shared hook repositories defined, those will live under the
`~/.githooks/shared` folder. The
[command line tool](https://github.com/gabyx/githooks/blob/main/docs/cli/git_hooks.md)
can list out all of them for you with `git hooks list`, and you can use it to
trust or ingoring hooks.

## More information

You can find more information about how this all works in the
[README](https://github.com/gabyx/githooks/blob/main/README.md) of the
[Githooks](https://github.com/gabyx/githooks) project repository.

If you find it useful, please show your support by starring the project in
GitHub!
