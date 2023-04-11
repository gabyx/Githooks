
# Runs the clean procedure.
clean:
	bash scripts/clean.sh

# Runs the build procedure.
apps:
	bash scripts/build.sh

# Test release
release-test:
	GORELEASER_CURRENT_TAG=v9.9.9 goreleaser release --snapshot --clean --skip-sign --skip-publish --skip-validate
.PHONY: apps clean
