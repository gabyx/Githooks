package git

const (
	// GitCKInitTemplateDir is the Git template dir config key.
	GitCKInitTemplateDir = "init.templateDir"

	// GitCKCoreHooksPath is the Git global hooks path config key.
	GitCKCoreHooksPath = "core.hooksPath"

	// GitCVTrue is the boolean `true` Git config value.
	GitCVTrue = "true"
	// GitCVFalse is the boolean `false` Git config value.
	GitCVFalse = "false"
)

// type GitConfigValue {
// 	key string
// 	value string
// 	changed bool
// }

// // GitConfigCache for faster read access.
// type GitConfigCache struct {
// 	local  map[string]GitConfigValue
// 	global map[string]GitConfigValue
// 	system map[string]GitConfigValue
// }

// func NewGitConfigCache() *GitConfigCache {
// 	CtxSanitized()
// }
