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

Variable `hookName` refers top one of the supported [Git hooks](/README.md).

### Version 1

```yaml
cmd: "/var/etc/lib/crazy/command"
args:
    - "--do-it"
    - "--faster"
    - "--proper"
version: 1
```

### Version 2

- Added environment variables `env`.

```yaml
cmd: "/var/etc/lib/crazy/command"
args:
    - "--do-it"
    - "--faster"
    - "--proper"
env:
    - USE_CUSTOM=1
version: 2
```
