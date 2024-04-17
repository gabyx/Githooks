package hooks

import (
	"fmt"
	"slices"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
)

func TestGithooksCompliesWithGit(t *testing.T) {
	doc, err := htmlquery.LoadURL("https://git-scm.com/docs/githooks")
	assert.Nil(t, err, "Could not load doc.")

	list := htmlquery.Find(doc, `//h2[@id="_hooks"]/following-sibling::div//h3`)
	assert.NotEmpty(t, list)

	var names []string
	for _, l := range list {
		names = append(names, l.LastChild.Data)
		assert.Contains(t, AllHookNames, l.LastChild.Data)
	}

	all := append([]string{}, AllHookNames...)
	slices.Sort(names)
	slices.Sort(all)

	fmt.Printf("Git hooks names: [%v]: %s\n", len(names), names)
	fmt.Printf("Git hooks names: [%v]: %s\n", len(all), all)

	assert.Equal(t, all, names, "Git contains not the same hooks")
}
