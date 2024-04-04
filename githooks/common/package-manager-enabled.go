//go:build !package_manager_enabled

package common

const (
	// We are using a package manager and
	// updating is forbidden and Githooks is
	// externally managed.
	PackageManagerEnabled = false
)
