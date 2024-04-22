package updates

import "github.com/gabyx/githooks/githooks/common"

const (
	// Only allow update if not using a package manager.
	Enabled = !common.PackageManagerEnabled
)
