//go:build mock || debug

package hooks

import (
	"github.com/gabyx/githooks/githooks/git"
)

const (
	UseThreadPool = true
)

func AllowLocalURLInRepoSharedHooks() bool {
	return git.NewCtx().GetConfig("githooks.testingTreatFileProtocolAsRemote", git.Traverse) == git.GitCVTrue
}
