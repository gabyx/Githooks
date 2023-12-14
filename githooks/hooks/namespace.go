package hooks

import (
	"os"
	"path"
	"strings"

	"github.com/agext/regexp"
	cm "github.com/gabyx/githooks/githooks/common"
)

const (
	// NamespaceRepositoryHook is the namespace for repository hooks.
	NamespaceRepositoryHook = "gh-self"
	// NamespaceReplacedHook is the namespace for replaced hooks.
	NamespaceReplacedHook = "gh-replaced"

	// NamespacePrefix prefixes the namespace part in the namespace path of a hook.
	NamespacePrefix = "ns:"
)

func getNamespaceFile(hooksDir string) string {
	return path.Join(hooksDir, ".namespace")
}

var sanitizeNamespace = regexp.MustCompile(`\s+|\/`)

// GetHooksNamespace get the namespace in which
// all hooks in `hooksDir` are residing.
func GetHooksNamespace(hookDir string) (s string, err error) {
	f := getNamespaceFile(hookDir)
	if cm.IsFile(f) {
		var data []byte
		data, err = os.ReadFile(f)
		s = sanitizeNamespace.ReplaceAllString(strings.TrimSpace(string(data)), "-")
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

// SplitNamespacePath splits a namespace path `ns:xxx/a/b/c` into
// `xxx` and `a/b/c`
func SplitNamespacePath(nsPath string) (ns, path string, err error) {
	s := strings.SplitN(nsPath, "/", 2)
	if len(s) > 1 {
		path = s[1]
	}

	ns = strings.TrimPrefix(s[0], NamespacePrefix)
	if ns == s[0] {
		err = cm.ErrorF("Namespace name '%s' does not have suffix 'ns:'.", nsPath)
		return
	}

	return
}
