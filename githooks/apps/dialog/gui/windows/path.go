//go:build windows

package gui

import "regexp"

var forbiddenFilenameCharsRe = regexp.MustCompile(`[<>:"\/\\|?*]`)

func SanitizeFilename(filename string) string {
	return forbiddenFilenameCharsRe.ReplaceAllString(filename, "-")
}
