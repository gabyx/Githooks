// +build darwin

package gui

import (
	"context"
	"strings"

	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
)

func ShowFileSave(ctx context.Context, s *set.FileSave) (res.File, error) {

	return res.File{}, nil
}

func ShowFileSelection(ctx context.Context, s *set.FileSelection) (res.File, error) {

	return res.File{}, nil
}

func initFilters(filters []set.FileFilter) []string {
	var res []string
	for _, f := range filters {
		var buf strings.Builder
		buf.WriteString("--file-filter=")
		if f.Name != "" {
			buf.WriteString(f.Name)
			buf.WriteRune('|')
		}
		for _, p := range f.Patterns {
			buf.WriteString(p)
			buf.WriteRune(' ')
		}
		res = append(res, buf.String())
	}

	return res
}
