package hooks

import (
	"io"
	"os"
	"testing"

	"github.com/gabyx/githooks/githooks/git"

	"github.com/stretchr/testify/assert"
)

func getGitConfig(key string, scope git.ConfigScope) (string, bool) {
	if key == "one.one" {
		return "", false
	}

	s := ""
	if scope == git.Traverse {
		s = "--traverse"
	} else {
		s = "--" + git.ToConfigName(scope)
	}

	return key + s, true
}

func TestEnvReplace(t *testing.T) {

	subst := getVarSubstitution(os.LookupEnv, getGitConfig)

	_ = os.Setenv("var", "banana")
	_ = os.Setenv("tar", "monkey")

	var r string
	var err error

	r, err = subst(`${var}`)
	assert.Equal(t, `${var}`, r, "Should not have been replaced.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`\${var}`)
	assert.Equal(t, `\${var}`, r, "Should not have been replaced.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`${env:gar}`)
	assert.Equal(t, "", r, "Replace non existent env. var.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`${env:var}`)
	assert.Equal(t, "banana", r, "Replace existent env. var.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`${env:var} ${env:tar}`)
	assert.Equal(t, "banana monkey", r, "Replace existent env. var.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`${git:one.one}`)
	assert.Equal(t, "", r, "Replace non existent Git var.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`${!git:one.one}`)
	assert.Equal(t, "", r, "Replace non existent Git var.")
	assert.NotNil(t, err, "Need an error.")

	r, err = subst(`${git:two}`)
	assert.Equal(t, "two--traverse", r, "Replace existent Git var.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`${git-l:two}`)
	assert.Equal(t, "two--local", r, "Replace existent Git var.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`${git-g:two}`)
	assert.Equal(t, "two--global", r, "Replace existent Git var.")
	assert.Nil(t, err, "No error.")

	r, err = subst(`${env:var} '${git-l:one.one}' ${git-l:two} ${git-g:two} ${git-s:two}`)
	assert.Equal(t, "banana '' two--local two--global two--system", r, "Replace existent Env and Git var.")
	assert.Nil(t, err, "No error.")

	// Test some error replacements
	r, err = subst(`'${git-l:one.one}' ${git-l:two} ${git-g:two} ${git-s:two} '${!env:nonexistentenvvar}'`)
	assert.Equal(t, "'' two--local two--global two--system ''", r, "Replace existent Env and Git var.")
	assert.NotNil(t, err, "Need an error.")

}

func TestRunnerConfigVersion(t *testing.T) {
	f, e := os.CreateTemp("", "")
	assert.Nil(t, e)

	defer func() { _ = os.Remove(f.Name()) }()
	_, e = io.WriteString(f,
		`
version: 999999
	  `)
	assert.Nil(t, e)

	_, e = loadRunnerConfig(f.Name())
	assert.Error(t, e)
	if e != nil {
		assert.Contains(t, e.Error(), "Githooks only supports version >= 1")
	}
}
