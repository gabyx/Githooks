package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookCountString(t *testing.T) {

	l := HookPrioList{[]Hook{Hook{}, Hook{}}, []Hook{Hook{}, Hook{}}, []Hook{Hook{}, Hook{}, Hook{}}}
	assert.Equal(t, l.CountString(), "[2,2,3]")

	l = HookPrioList{}
	assert.Equal(t, l.CountString(), "[0]")
}
