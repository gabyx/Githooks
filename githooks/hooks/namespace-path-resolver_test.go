package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaceSplit(t *testing.T) {
	ns, path, err := SplitNamespacePath("ns:gh-self/a/b/c")
	assert.NoError(t, err)
	assert.Equal(t, "gh-self", ns)
	assert.Equal(t, "a/b/c", path)

	ns, path, err = SplitNamespacePath("ns:a")
	assert.NoError(t, err)
	assert.Equal(t, "a", ns)
	assert.Empty(t, path)

	_, _, err = SplitNamespacePath("gh-self/a/b/c")
	assert.Error(t, err)
}
