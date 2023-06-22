# YAML Specification

## Ignore File `.ignore.yaml`

### Version 1

```yaml
patterns:
  - "my-super-shared-hooks/**/.sh"
  - "!my-super-shared-hooks/pre-commit/*.sh"
  - "**/*.md"
  - "pre/-commit/*.py"

paths:
  - "commit-msg/*check*"
  - "hooks/pre-commit.replaced.githook"

version: 1
```

## Shared Hooks Configuration `.shared.yaml`

### Version 1

```yaml
urls:
  - "ssh://github.com/shared/hooks-go.git@mybranch"
  - "git@github.com:shared/hooks-maven.git"
  - "git://github.com/shared/hooks-python.git"
  - "file:///local/path/to/bare-repo.git@mybranch"

version: 1
```

## Hook Run Configuration `<hookName>.yaml`

Variable `hookName` refers to one of the supported [Git hooks](/README.md).

### Version 1

```yaml
cmd: "/var/etc/lib/crazy/command"
args: # optional
  - "--do-it"
  - "--faster"
  - "--all"
  - "${env:GPG_PUBLIC_KEY}"
  - "--test ${git-l:my-local-git-config-var}"
version: 1 # optional
```

### Version 2

- Added environment variables `env`.

```yaml
cmd: "/var/etc/lib/crazy/command"
args: # optional
  - "--do-it"
  - "${env:GPG_PUBLIC_KEY}"
  - "--test ${git-l:my-local-git-config-var}"
env: # optional
  - USE_CUSTOM=1
version: 2 # optional
```

### Version 3

- Added image field `image`.

```yaml
cmd: "/var/etc/lib/crazy/command"
args: # optional
  - "--do-it"
  - "${env:GPG_PUBLIC_KEY}"
  - "--test ${git-l:my-local-git-config-var}"
env: # optional
  - USE_CUSTOM=1
image: # optional
  reference: mycontainerimage:1.2.0
version: 3 # optional
```
