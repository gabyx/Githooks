//go:build !windows

package installer

// GetDefaultTemplateSearchDir returns the search directories for potential template dirs.
func GetDefaultTemplateSearchDir() ([]string, []string) {
	return []string{"/usr"}, []string{"/"}
}
