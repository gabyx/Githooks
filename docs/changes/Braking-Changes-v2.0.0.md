# API Breaking Changes v2.0.0: Namespace Path & Ingore Specification

Pull request [#35](https://github.com/gabyx/Githooks/pull/35) introduces a breaking change in the
definition of a hook's *namespace path*.
The original definition is to little explicit which poses problems in properly distinguising
patterns which contain a namespace from ones which do not.

The following changes are made:

- Before the namespace of a hook was simply prefixed by `<hookNamespace>/...`, e.g. examples are

    - `myhooks/pre-commit/format.sh` for a shared hook repo with `<hooksDir>/.namespace` containing `myhooks`.
    - `pre-commit/format.sh` for a repository hook in `<repoDir>/.githooks` where Githooks is running.

  New namespace paths (see documentation) are written with a explicit suffix `ns:` like:

    - `ns:myhooks/pre-commit/format.sh` for a shared hook repo with `myhooks` in `<hooksDir>/.namespace`.
    - `ns:gh-self/pre-commit/format.sh` for a repository hook in `<repoDir>/.githooks` where Githooks
      is running or similar

    - `ns:myhooks/pre-commit/format.sh` for a repository hook in `<repoDir>/.githooks` where Githooks
      is running where `<repoDir>/.githooks/.namespace` contains `myhooks`.

- Furthermore, ignore patterns in `.ignore.yaml` which contain no `ns:` at the beginning are now considered
  relative patterns. Before they matched the whole namespace path of a hook, including the namespace itself.
  Now these patterns are relative and are made absolute to the namespace of the current `<hooksDir>`, e.g. a pattern

  - `**/pre-commit/format.sh` in `.githooks/.ignore.yaml` will be translated into `ns:gh-self/**/pre-commit/format.sh`
    and a pattern
  - `**/batch*/format.sh` in `.githooks/pre-commit/.ignore.yaml` will be translated into
    `ns:gh-self/pre-commit/**/batch*/format.sh` where `gh-self` is the internal default namespace
    for the repository where Githooks is running.

- Replaced hooks are now addressed by a namespace `gh-replaced`, e.g. `ns:gh-replaced/pre-commit.replaced.githooks.`

## Upgrade Resolution

- Check `git hooks list` and adapt the `.ignore.yaml` files to adapt to the new namespace path.
