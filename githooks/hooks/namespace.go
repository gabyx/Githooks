package hooks

import (
	"os"
	"path"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
)

const (
	// NamespaceRepositoryHook is the namespace for repository hooks.
	NamespaceRepositoryHook = ""
	// NamespaceReplacedHook is the namespace for replace hooks.
	NamespaceReplacedHook = "hooks"
)

func getNamespaceFile(hooksDir string) string {
	return path.Join(hooksDir, ".namespace")
}

// GetHooksNamespace get the namespace in which
// all hooks in `hooksDir` are residing.
func GetHooksNamespace(hookDir string) (s string, err error) {
	f := getNamespaceFile(hookDir)
	if cm.IsFile(f) {
		var data []byte
		data, err = os.ReadFile(f)
		s = strings.TrimSpace(string(data))
	}

	return
}

// GetDefaultHooksNamespaceShared returns the default hooks namespace for
// a shared url.
func GetDefaultHooksNamespaceShared(sharedRepo *SharedRepo) string {
	hash, err := cm.GetSHA1Hash(strings.NewReader(sharedRepo.OriginalURL))
	cm.AssertNoErrorPanic(err, "Could not compute default hash.")

	return hash[0:10]
}
