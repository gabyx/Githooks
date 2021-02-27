package common

import (
	"gabyx/githooks/apps/dialog/settings"
	strs "gabyx/githooks/strings"
	"strings"
)

type fileFilterArgs struct {
	Filters *[]settings.FileFilter
}

func (f *fileFilterArgs) String() string {
	return strs.Fmt("%v", f.Filters)
}

func (f *fileFilterArgs) Type() string {
	return "[]FileFilter"
}

func (f *fileFilterArgs) Set(s string) error {
	var filter settings.FileFilter

	if split := strings.SplitN(s, "|", 2); len(split) > 1 {
		filter.Name = split[0]
		s = split[1]
	}

	filter.Patterns = strings.Split(strings.TrimSpace(s), " ")
	*f.Filters = append(*f.Filters, filter)

	return nil
}
