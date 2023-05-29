package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookCountString(t *testing.T) {

	l := HookPrioList{[]Hook{{}, {}}, []Hook{{}, {}}, []Hook{{}, {}, {}}}
	assert.Equal(t, l.CountFmt(), "[2,2,3]")

	l = HookPrioList{}
	assert.Equal(t, l.CountFmt(), "[0]")
}
