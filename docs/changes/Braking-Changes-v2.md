# API Breaking Changes v2.x.x

## Namespace Path & Ignore Specification

Pull request [#35](https://github.com/gabyx/Githooks/pull/35) introduces a breaking change in the
definition of a hook's *namespace path*.
The original definition is not explicit enought which poses problems for properly distinguishing
patterns which contain a namespace from ones which don't.

The following changes are made:

- Before, the namespace of a hook was simply prefixed by `<hookNamespace>/...`, e.g. examples were

    - `myhooks/pre-commit/format.sh` for a shared hook repo with `<hooksDir>/.namespace` containing `myhooks`.
    - `pre-commit/format.sh` for a repository hook in `<repoDir>/.githooks` where Githooks is running.

  Now, new namespace paths (see documentation) are written with a explicit suffix `ns:` like:

    - `ns:myhooks/pre-commit/format.sh` for a shared hook repository with `myhooks` in `<hooksDir>/.namespace`.
    - `ns:gh-self/pre-commit/format.sh` for a repository hook in `<repoDir>/.githooks` where Githooks
      is running.

    - `ns:myhooks/pre-commit/format.sh` for a repository hook in `<repoDir>/.githooks` where Githooks
      is running where `<repoDir>/.githooks/.namespace` contains `myhooks`.

- Furthermore, ignore patterns in `.ignore.yaml` which contain no prefix `ns:` are now considered
  relative patterns. Before they matched the whole namespace path of a hook, including the namespace itself.
  Now, these patterns are relative and are made absolute to the namespace of the current `<hooksDir>`, e.g. a pattern

  - `**/pre-commit/format.sh` in `.githooks/.ignore.yaml` will be the same as `ns:gh-self/**/pre-commit/format.sh`
    and a pattern
  - `**/batch*/format.sh` in `.githooks/pre-commit/.ignore.yaml` will be the same as
    `ns:gh-self/pre-commit/**/batch*/format.sh` where `gh-self` is the internal default namespace
    for the repository where Githooks is running.

- Replaced hooks are now addressed by a namespace `gh-replaced`, e.g. `ns:gh-replaced/pre-commit.replaced.githook`

## Upgrade Resolution

- Check `git hooks list` and adapt the `.ignore.yaml` files to adapt to the new namespace path.
