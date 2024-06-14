package common

import (
	"strconv"

	cm "github.com/gabyx/githooks/githooks/common"
)

type indexArgs struct {
	indices *[]uint
}

func (i *indexArgs) String() string {
	return ""
}

func (i *indexArgs) Type() string {
	return "[]uint"
}

func (i *indexArgs) Set(s string) error {

	value, err := strconv.ParseInt(s, 10, 32) // nolint: mnd
	if err != nil {
		return cm.ErrorF("Could not parse index '%s'.", s)
	}

	*i.indices = append(*i.indices, uint(value))

	return nil
}
