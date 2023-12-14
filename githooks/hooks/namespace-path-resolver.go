package hooks

import (
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
)

type QueryResult struct {
	Namespace      string
	NamespacePath  string
	RepositoryRoot string
	HooksDir       string
	Found          bool
}

func ResolveNamespacePaths(
	log cm.ILogContext,
	gitx *git.Context,
	installDir string,
	repoDir string,
	nsPaths []string) (res []QueryResult, foundAll bool, err error) {

	res = make([]QueryResult, len(nsPaths))

	for i := range nsPaths {
		res[i].Namespace, res[i].NamespacePath, err = SplitNamespacePath(nsPaths[i])
		if err != nil {
			return
		}
	}

	localHooksDir := GetGithooksDir(repoDir)

	// Cycle through all shared hooks an return the first with matching namespace.
	allRepos, err := LoadRepoSharedHooks(installDir, repoDir)
	log.AssertNoErrorPanicF(err, "Could not load shared hook list '%s'.", GetRepoSharedFileRel())
	local, err := LoadConfigSharedHooks(installDir, gitx, git.LocalScope)
	log.AssertNoErrorPanicF(err, "Could not load local shared hook list.")
	global, err := LoadConfigSharedHooks(installDir, gitx, git.GlobalScope)
	log.AssertNoErrorPanicF(err, "Could not load local shared hook list.")

	allRepos = append(allRepos, local...)
	allRepos = append(allRepos, global...)
	found := 0

	for rI := range allRepos {
		if !cm.IsDirectory(allRepos[rI].RepositoryDir) {
			continue
		}

		hooksDir := GetSharedGithooksDir(allRepos[rI].RepositoryDir)
		ns, err := GetHooksNamespace(hooksDir)
		log.AssertNoErrorPanicF(err, "Could not get hook namespace in '%s'", hooksDir)

		for nI := range res {
			p := &res[nI]

			switch {
			case p.Found:
				continue
			case p.Namespace == NamespaceRepositoryHook:
				p.RepositoryRoot = repoDir
				p.HooksDir = localHooksDir

				p.Found = true
				found++
			case p.Namespace == ns:
				p.RepositoryRoot = allRepos[rI].RepositoryDir
				p.HooksDir = hooksDir

				p.Found = true
				found++
			}
		}
	}

	if found == len(res) {
		foundAll = true
	}

	return
}
