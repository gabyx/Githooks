package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookCountString(t *testing.T) {
	l := HookPrioList{[]Hook{{}, {}}, []Hook{{}, {}}, []Hook{{}, {}, {}}}
	assert.Equal(t, "[2,2,3]", l.CountFmt())

	l = HookPrioList{}
	assert.Equal(t, "[0]", l.CountFmt())
}
