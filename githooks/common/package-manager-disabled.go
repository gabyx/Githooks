//go:build package_manager_enabled

package common

const (
	// We are not using a package manager and
	// updating is allwed and Githooks is
	// not externally managed.
	PackageManagerEnabled = true
)
