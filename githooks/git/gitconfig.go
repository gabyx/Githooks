package git

const (
	// If you add/remove settings here, you must edit `hooks.filterRegex`.

	// GitCKInitTemplateDir is the Git template dir config key.
	GitCKInitTemplateDir = "init.templateDir"

	// GitCKCoreHooksPath is the Git global hooks path config key.
	GitCKCoreHooksPath = "core.hooksPath"

	// GitCVTrue is the boolean `true` Git config value.
	GitCVTrue = "true"
	// GitCVFalse is the boolean `false` Git config value.
	GitCVFalse = "false"
)
