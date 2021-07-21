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

	// NamespaceSeparator separates the namespace from the rest in the namespace path of a hook.
	// Since it contains `/` its impossible to use in file names an
	// thus is safe to check for in ignore patterns.
	NamespaceSeparator = "://"
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
		s = strings.ReplaceAll(strings.TrimSpace(string(data)), " ", "-")
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
