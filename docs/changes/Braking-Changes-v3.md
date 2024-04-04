# API Breaking Changes v3.x.x to v2

## Githooks Updates

Githooks does not run updates automatically anymore. Update checks are however
still executed if `git hooks config enable-update-checks` is enabled. The
following settings are changed:

`githooks.autoUpdateEnabled` -> `githooks.updateCheckEnabled`
`githooks.autoUpdateUsePrerelease` -> `githooks.updateCheckUsePrerelease`

Any upgrade to `v3` will change this setting automatically.
