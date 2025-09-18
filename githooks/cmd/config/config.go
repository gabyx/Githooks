package config

import (
	"strings"
	"time"

	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	"github.com/gabyx/githooks/githooks/cmd/disable"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
	"github.com/gabyx/githooks/githooks/updates"

	"github.com/pkg/math"
	"github.com/spf13/cobra"
)

// GitOptions for local or global Git configuration.
type GitOptions struct {
	Local  bool
	Global bool
}

func wrapToGitScope(log cm.ILogContext, opts *GitOptions) git.ConfigScope {
	switch {
	default:
		fallthrough
	case opts.Local && opts.Global:
		log.PanicF("You cannot use '--local' or '--global' at the same time.")

		return git.LocalScope
	case opts.Local:
		return git.LocalScope
	case opts.Global:
		return git.GlobalScope
	}
}

// SetOptions holds data for setting Git options to either
// print/reset/unset/set a value.
type SetOptions struct {
	Print bool
	Reset bool

	Unset  bool
	Set    bool
	Values []string
}

// AssertOptions  asserts that all set actions are exclusive etc.
func (s *SetOptions) AssertOptions(log cm.ILogContext, optsMap *OptionsMapping, noValues bool, args []string) {

	log.PanicIf(!s.Set && !s.Unset && !s.Reset && !s.Print, "You need to specify an option.")

	log.PanicIfF(s.Print && (s.Reset || s.Unset || s.Set || len(args) != 0),
		"You cannot use '--%s' with any other options\n"+ // nolint: goconst
			"or arguments at the same time.", optsMap.Print) // nolint: goconst

	log.PanicIfF(s.Reset && (s.Unset || s.Print || s.Set || len(args) != 0),
		"You cannot use '--%s' with any other options\n"+
			"or arguments at the same time.", optsMap.Reset)

	log.PanicIfF(s.Unset && (s.Print || s.Reset || s.Set || len(args) != 0),
		"You cannot use '--%s' with any other options\n"+
			"or arguments at the same time.", optsMap.Unset)

	log.PanicIfF(s.Set && (s.Print || s.Reset || s.Unset || (!noValues && len(args) == 0)),
		"You cannot use '--%s' with any other options\n"+
			"and you need to specify values.", optsMap.Set)

	if s.Set {
		for i := range args {
			log.PanicIfF(strs.IsEmpty(args[i]), "Argument '%v' may not be empty.", args[i])
		}

		s.Values = args
	}
}

// OptionsMapping holds mappings for the
// print/set/unset/reset actions.
type OptionsMapping struct {
	Print     string
	PrintDesc string
	Set       string
	SetDesc   string
	Unset     string
	UnsetDesc string
	Reset     string
	ResetDesc string
}

func createOptionMap(hasSet bool, hasUnset bool, hasReset bool) OptionsMapping {
	opts := OptionsMapping{
		Print:     "print",
		PrintDesc: "Print the setting."}

	if hasSet {
		opts.Set = "set"
		opts.SetDesc = "Set the setting."
	}

	if hasUnset {
		opts.Unset = "unset"
		opts.UnsetDesc = "Unset the setting."
	}

	if hasReset {
		opts.Reset = "reset"
		opts.ResetDesc = "Reset the setting."
	}

	return opts
}

func wrapToEnableDisable(opts *OptionsMapping) {
	opts.Set = "enable"
	opts.Unset = "disable"
}

func configSetOptions(
	cmd *cobra.Command,
	opts *SetOptions,
	optsMap *OptionsMapping,
	log cm.ILogContext,
	nMinArgs int, nMaxArgs int) {

	if strs.IsNotEmpty(optsMap.Print) {
		cmd.Flags().BoolVar(&opts.Print, optsMap.Print, false, optsMap.PrintDesc)
	}
	if strs.IsNotEmpty(optsMap.Set) {
		cmd.Flags().BoolVar(&opts.Set, optsMap.Set, false, optsMap.SetDesc)
	}
	if strs.IsNotEmpty(optsMap.Unset) {
		cmd.Flags().BoolVar(&opts.Unset, optsMap.Unset, false, optsMap.UnsetDesc)
	}
	if strs.IsNotEmpty(optsMap.Reset) {
		cmd.Flags().BoolVar(&opts.Reset, optsMap.Reset, false, optsMap.ResetDesc)
	}

	rangeCheck := ccm.PanicIfNotRangeArgs(log, nMinArgs, nMaxArgs)
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		opts.AssertOptions(log, optsMap, nMaxArgs == 0, args)
		if opts.Set {
			rangeCheck(cmd, args)
		}
	}
}

func runList(ctx *ccm.CmdContext, gitOpts *GitOptions) {

	print := func(scope git.ConfigScope) string {

		pairs := ctx.GitX.GetConfigRegex("(^githooks|alias.hooks)", scope)

		maxLength := 0
		for i := range pairs {
			maxLength = math.MaxInt(maxLength, len(pairs[i].Key)+2) // nolint: mnd
		}
		keyFmt := strs.Fmt("%%-%vs", maxLength)

		if len(pairs) == 0 {
			return "[0]: none"
		}

		var sb strings.Builder
		_, err := strs.FmtW(&sb, "[%v]:", len(pairs))
		cm.AssertNoErrorPanic(err, "Could not write message.")

		for i := range pairs {
			key := strs.Fmt("'%s'", pairs[i].Key)
			_, err = strs.FmtW(&sb, "\n%s "+keyFmt+" : '%s'", cm.ListItemLiteral, key, pairs[i].Value)
			cm.AssertNoErrorPanic(err, "Could not write message.")
		}

		return sb.String()
	}

	if gitOpts.Local {
		ctx.Log.InfoF("Local Githooks configurations %s", print(git.LocalScope))
	}

	if gitOpts.Global {
		ctx.Log.InfoF("Global Githooks configurations %s", print(git.GlobalScope))
	}

}

func runDisable(ctx *ccm.CmdContext, opts *SetOptions, gitOpts *GitOptions) {
	disable.RunDisable(ctx, opts.Reset, opts.Print, gitOpts.Global)
}

func runSearchDir(ctx *ccm.CmdContext, opts *SetOptions) {
	opt := hooks.GitCKPreviousSearchDir
	switch {
	case opts.Set:
		err := ctx.GitX.SetConfig(opt, opts.Values[0], git.GlobalScope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set Git config '%s'.", opt)
		ctx.Log.InfoF("Set previous search directory used during install to\n'%s'.", opts.Values[0])

	case opts.Reset:
		err := ctx.GitX.UnsetConfig(opt, git.GlobalScope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not unset Git config '%s'.", opt)
		ctx.Log.Info("Unset previous search directory used during install.")

	case opts.Print:
		conf := ctx.GitX.GetConfig(opt, git.GlobalScope)
		if strs.IsEmpty(conf) {
			ctx.Log.InfoF("Previous search directory is not set.")
		} else {
			ctx.Log.InfoF("Previous search directory is set to\n'%s'.", conf)
		}
	default:
		cm.Panic("Wrong arguments.")
	}
}

func runContainerManagerTypes(ctx *ccm.CmdContext, opts *SetOptions, gitOpts *GitOptions) {
	opt := hooks.GitCKContainerManager
	localOrGlobal := "locally" // nolint: goconst
	if gitOpts.Global {
		localOrGlobal = "globally" // nolint: goconst
	}

	scope := wrapToGitScope(ctx.Log, gitOpts)
	switch {
	case opts.Set:
		val := strings.Join(opts.Values, ",")
		err := ctx.GitX.SetConfig(opt, val, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set Git config '%s'.", opt)
		ctx.Log.InfoF("Container manager types is set to '%s' %s.", val, localOrGlobal)

	case opts.Reset:
		err := ctx.GitX.UnsetConfig(opt, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not unset Git config '%s'.", opt)
		ctx.Log.InfoF("Container manager types is unset %s.", localOrGlobal)

	case opts.Print:
		conf := ctx.GitX.GetConfig(opt, scope)
		ctx.Log.InfoF("Container manager types is set to '%s' %s.", conf, localOrGlobal)
	default:
		cm.Panic("Wrong arguments.")
	}
}

func runContainerizedHooksEnable(ctx *ccm.CmdContext, opts *SetOptions, gitOpts *GitOptions) {
	opt := hooks.GitCKContainerizedHooksEnabled
	localOrGlobal := "locally" // nolint: goconst
	if gitOpts.Global {
		localOrGlobal = "globally" // nolint: goconst
	}

	scope := wrapToGitScope(ctx.Log, gitOpts)
	switch {
	case opts.Set:
		err := ctx.GitX.SetConfig(opt, git.GitCVTrue, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set Git config '%s'.", opt)
		ctx.Log.InfoF("Running hooks containerized is now enabled %s.", localOrGlobal)

	case opts.Reset:
		err := ctx.GitX.UnsetConfig(opt, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not unset Git config '%s'.", opt)
		ctx.Log.InfoF("Running hooks containerized is now disabled %s.", localOrGlobal)

	case opts.Print:
		conf := ctx.GitX.GetConfig(opt, scope)
		if conf == git.GitCVTrue {
			ctx.Log.InfoF("Running hooks containerized is enabled %s.", localOrGlobal)
		} else {
			ctx.Log.InfoF("Running hooks containerized is disabled %s.", localOrGlobal)
		}
	default:
		cm.Panic("Wrong arguments.")
	}
}

func runSharedRepos(ctx *ccm.CmdContext, opts *SetOptions, gitOpts *GitOptions) {
	opt := hooks.GitCKShared

	localOrGlobal := "local"
	if gitOpts.Global {
		localOrGlobal = "global"
	}

	switch {
	case opts.Set:
		scope := wrapToGitScope(ctx.Log, gitOpts)
		for i := range opts.Values {
			err := ctx.GitX.AddConfig(opt, opts.Values[i], scope)
			ctx.Log.AssertNoErrorPanicF(err, "Could not add %s shared repository.", localOrGlobal)
		}
		ctx.Log.InfoF("Added '%v' %s shared repositories.", len(opts.Values), localOrGlobal)

	case opts.Reset:
		scope := wrapToGitScope(ctx.Log, gitOpts)
		err := ctx.GitX.UnsetConfig(opt, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not unset %s shared repository.", localOrGlobal)
		ctx.Log.InfoF("Removed all %s shared repositories.", localOrGlobal)

	case opts.Print:
		list := func(sh []string) string {
			if len(sh) == 0 {
				return "[0]: none"
			}

			return strs.Fmt("[%v]:\n%s", len(sh),
				strings.Join(strs.Map(sh,
					func(s string) string { return strs.Fmt("%s '%s'", cm.ListItemLiteral, s) }),
					"\n"))
		}

		if gitOpts.Local {
			shared := ctx.GitX.GetConfigAll(opt, git.LocalScope)
			ctx.Log.InfoF("Local shared repositories %s", list(shared))
		}

		if gitOpts.Global {
			shared := ctx.GitX.GetConfigAll(opt, git.GlobalScope)
			ctx.Log.InfoF("Global shared repositories %s", list(shared))
		}

	default:
		cm.Panic("Wrong arguments.")
	}
}

func runCloneURL(ctx *ccm.CmdContext, opts *SetOptions) {
	switch {
	case opts.Set:
		err := updates.SetCloneURL(opts.Values[0], "")
		ctx.Log.AssertNoErrorPanic(err, "Could not set Git hooks clone url.")
		ctx.Log.InfoF("Set Githooks clone URL to '%s'.", opts.Values[0])

	case opts.Print:
		url, branch := updates.GetCloneURL(ctx.GitX)
		if strs.IsNotEmpty(branch) {
			ctx.Log.InfoF("Githooks clone URL is set to '%s' at branch '%s'.", url, branch)
		} else {
			ctx.Log.InfoF("Githooks clone URL is set to '%s' at default branch.", url)
		}
	default:
		cm.Panic("Wrong arguments.")
	}
}

func runCloneBranch(ctx *ccm.CmdContext, opts *SetOptions) {

	switch {
	case opts.Set:
		err := updates.SetCloneBranch(opts.Values[0])
		ctx.Log.AssertNoErrorPanic(err, "Could not set Git hooks clone branch.")
		ctx.Log.InfoF("Set Githooks clone branch to '%s'.", opts.Values[0])

	case opts.Reset:
		err := updates.ResetCloneBranch()
		ctx.Log.AssertNoErrorPanic(err, "Could not unset Git hooks clone branch.")
		ctx.Log.Info("Unset Githooks clone branch. Using default branch.")

	case opts.Print:
		url, branch := updates.GetCloneURL(ctx.GitX)
		if strs.IsNotEmpty(branch) {
			ctx.Log.InfoF("Githooks clone URL is set to '%s' at branch '%s'.", url, branch)
		} else {
			ctx.Log.InfoF("Githooks clone URL is set to '%s' at default branch.", url)
		}
	default:
		cm.Panic("Wrong arguments.")
	}
}

func runUpdateTime(ctx *ccm.CmdContext, opts *SetOptions) {
	const text = "Githooks update check timestamp"

	switch {
	case opts.Reset:
		err := updates.ResetUpdateCheckTimestamp(ctx.InstallDir)
		ctx.Log.AssertNoErrorPanicF(err, "Could not reset %s.", text)
		ctx.Log.InfoF("Reset %s.", text)

	case opts.Print:
		ts, isSet, err := updates.GetUpdateCheckTimestamp(ctx.InstallDir)
		ctx.Log.AssertNoErrorPanic(err, "Could not get %s.", text)
		if isSet {
			ctx.Log.InfoF("%s set to '%s'.", ts.Format(time.RFC1123))
		} else {
			ctx.Log.InfoF("%s is not set.\n"+
				"Update checks have never run.", text)
		}
	default:
		cm.Panic("Wrong arguments.")
	}
}

func runTrustAllHooks(ctx *ccm.CmdContext, opts *SetOptions) {

	ccm.AssertRepoRoot(ctx)

	switch {
	case opts.Set:
		err := hooks.SetTrustAllSetting(ctx.GitX, true, false)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set trust-all-hooks setting.")
		ctx.Log.InfoF("The current repository trusts all hooks automatically.")

	case opts.Unset:
		err := hooks.SetTrustAllSetting(ctx.GitX, false, false)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set trust-all-hooks  setting.")
		ctx.Log.InfoF("The current repository trusts all hooks automatically.")

	case opts.Reset:
		err := hooks.SetTrustAllSetting(ctx.GitX, false, true)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set trust-all setting.")
		ctx.Log.InfoF("The trust-all-hooks setting is not set in the current repository.")

	case opts.Print:
		enabled, isSet := hooks.GetTrustAllSetting(ctx.GitX)
		switch {
		case !isSet:
			ctx.Log.Info("The trust-all-hooks setting is not set in the current repository.")
		case enabled:
			ctx.Log.Info("The current repository trusts all hooks automatically.")
		default:
			ctx.Log.Info("The current repository does not trust hooks automatically.")
		}

	default:
		cm.Panic("Wrong arguments.")
	}

}

// RunUpdateCheck enables/disables Githooks update checks.
func RunUpdateCheck(ctx *ccm.CmdContext, opts *SetOptions) {
	const text = "Githooks update"

	switch {
	case opts.Set:
		err := updates.SetUpdateCheckSettings(true, false)
		ctx.Log.AssertNoErrorPanicF(err, "Could not enable update settings.")
		ctx.Log.InfoF("%s checks are now enabled.", text)

	case opts.Unset:
		err := updates.SetUpdateCheckSettings(false, false)
		ctx.Log.AssertNoErrorPanicF(err, "Could not disable update settings.")
		ctx.Log.InfoF("%s checks are now disabled.", text)

	case opts.Reset:
		err := updates.SetUpdateCheckSettings(false, true)
		ctx.Log.AssertNoErrorPanicF(err, "Could not reset update settings.")
		ctx.Log.InfoF("%s setting is now unset.", text)

	case opts.Print:
		enabled, _ := updates.GetUpdateCheckSettings(ctx.GitX)
		switch {
		case enabled:
			ctx.Log.InfoF("%s checks are enabled.", text)
		default:
			ctx.Log.InfoF("%s checks are disabled.", text)
		}

	default:
		cm.Panic("Wrong arguments.")
	}
}

func runRunnerNonInteractive(ctx *ccm.CmdContext, opts *SetOptions, gitOpts *GitOptions) {
	scope := wrapToGitScope(ctx.Log, gitOpts)

	localOrGlobal := "locally" //nolint: goconst
	if gitOpts.Global {
		localOrGlobal = "globally" //nolint: goconst
	}

	const text = "non-interactive runner mode"
	switch {
	case opts.Set:
		err := hooks.SetRunnerNonInteractive(ctx.GitX, true, false, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not enable %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Enabled %s %s.", text, localOrGlobal)

	case opts.Unset:
		err := hooks.SetRunnerNonInteractive(ctx.GitX, false, false, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not disable %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Disabled %s %s.", text, localOrGlobal)

	case opts.Reset:
		err := hooks.SetRunnerNonInteractive(ctx.GitX, false, true, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not reset %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Reset %s %s.", text, localOrGlobal)

	case opts.Print:
		localOrGlobal = " " + localOrGlobal
		if !gitOpts.Global && !gitOpts.Local {
			scope = git.Traverse
			localOrGlobal = ""
		}

		enabled := hooks.IsRunnerNonInteractive(ctx.GitX, scope)
		if enabled {
			ctx.Log.InfoF("Non-interactive runner mode is enabled%s.", text, localOrGlobal)
		} else {
			ctx.Log.InfoF("Non-interactive runner mode is disabled%s.", text, localOrGlobal)
		}

	default:
		cm.Panic("Wrong arguments.")
	}
}

func runSkipNonExistingSharedHooks(ctx *ccm.CmdContext, opts *SetOptions, gitOpts *GitOptions) {
	scope := wrapToGitScope(ctx.Log, gitOpts)

	localOrGlobal := "locally" //nolint: goconst
	if gitOpts.Global {
		localOrGlobal = "globally" //nolint: goconst
	}

	const text = "non-existing shared hooks"
	switch {
	case opts.Set:
		err := hooks.SetSkipNonExistingSharedHooks(ctx.GitX, true, false, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not enable skipping %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Enabled skipping %s %s.", text, localOrGlobal)

	case opts.Unset:
		err := hooks.SetSkipNonExistingSharedHooks(ctx.GitX, false, false, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not disable skipping %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Disabled skipping %s %s.", text, localOrGlobal)

	case opts.Reset:
		err := hooks.SetSkipNonExistingSharedHooks(ctx.GitX, false, true, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not reset skipping %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Reset skipping %s %s.", text, localOrGlobal)

	case opts.Print:
		localOrGlobal = " " + localOrGlobal
		if !gitOpts.Global && !gitOpts.Local {
			scope = git.Traverse
			localOrGlobal = ""
		}

		enabled := hooks.SkipNonExistingSharedHooks(ctx.GitX, scope)
		if enabled {
			ctx.Log.InfoF("Skipping %s is enabled%s.", text, localOrGlobal)
		} else {
			ctx.Log.InfoF("Skipping %s is disabled%s.", text, localOrGlobal)
		}

	default:
		cm.Panic("Wrong arguments.")
	}
}

func runDisableSharedHooksUpdate(ctx *ccm.CmdContext, opts *SetOptions, gitOpts *GitOptions) {
	scope := wrapToGitScope(ctx.Log, gitOpts)

	localOrGlobal := "locally" //nolint: goconst
	if gitOpts.Global {
		localOrGlobal = "globally" //nolint: goconst
	}

	const text = "automatic updates of shared hooks"
	switch {
	case opts.Set:
		err := hooks.SetDisableSharedHooksUpdate(ctx.GitX, true, false, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not disable %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Disable %s %s.", text, localOrGlobal)

	case opts.Unset:
		err := hooks.SetDisableSharedHooksUpdate(ctx.GitX, false, false, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not enable %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Enabled %s %s.", text, localOrGlobal)

	case opts.Reset:
		err := hooks.SetDisableSharedHooksUpdate(ctx.GitX, false, true, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not reset (enable) %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Reset (enabled) %s %s.", text, localOrGlobal)

	case opts.Print:
		localOrGlobal = " " + localOrGlobal
		if !gitOpts.Global && !gitOpts.Local {
			scope = git.Traverse
			localOrGlobal = ""
		}

		msg := "Automatic updates of shared hooks are "
		disabled, _ := hooks.IsSharedHooksUpdateDisabled(ctx.GitX, scope)
		if disabled {
			ctx.Log.InfoF(msg+"disabled%s.", text, localOrGlobal)
		} else {
			ctx.Log.InfoF(msg+"enabled%s.", text, localOrGlobal)
		}

	default:
		cm.Panic("Wrong arguments.")
	}
}

func runSkipUntrustedHooks(ctx *ccm.CmdContext, opts *SetOptions, gitOpts *GitOptions) {
	scope := wrapToGitScope(ctx.Log, gitOpts)

	localOrGlobal := "locally" //nolint: goconst
	if gitOpts.Global {
		localOrGlobal = "globally" //nolint: goconst
	}

	const text = "active, untrusted hooks"
	switch {
	case opts.Set:
		err := hooks.SetSkipUntrustedHooks(ctx.GitX, true, false, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not enable skipping %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Enabled skipping %s %s.", text, localOrGlobal)

	case opts.Unset:
		err := hooks.SetSkipUntrustedHooks(ctx.GitX, false, false, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not disable skipping %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Disabled skipping %s %s.", text, localOrGlobal)

	case opts.Reset:
		err := hooks.SetSkipUntrustedHooks(ctx.GitX, false, true, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not reset skipping %s %s.", text, localOrGlobal)
		ctx.Log.InfoF("Reset skipping %s %s.", text, localOrGlobal)

	case opts.Print:
		localOrGlobal = " " + localOrGlobal
		if !gitOpts.Global && !gitOpts.Local {
			scope = git.Traverse
			localOrGlobal = ""
		}

		enabled, _ := hooks.SkipUntrustedHooks(ctx.GitX, scope)
		if enabled {
			ctx.Log.InfoF("Skipping %s is enabled%s.", text, localOrGlobal)
		} else {
			ctx.Log.InfoF("Skipping %s is disabled%s.", text, localOrGlobal)
		}

	default:
		cm.Panic("Wrong arguments.")
	}
}

func runDeleteDetectedLFSHooks(ctx *ccm.CmdContext, opts *SetOptions) {
	opt := hooks.GitCKDeleteDetectedLFSHooksAnswer

	switch {
	case opts.Set:
		err := ctx.GitX.SetConfig(opt, "y", git.GlobalScope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set Git config '%s'.", opt)
		ctx.Log.InfoF("Detected LFS hooks will now automatically be deleted during install.")

	case opts.Unset:
		err := ctx.GitX.SetConfig(opt, "n", git.GlobalScope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set Git config '%s'.", opt)
		ctx.Log.Info("Detected LFS hooks will now automatically be deleted during install",
			"but instead backed up.")

	case opts.Reset:
		err := ctx.GitX.UnsetConfig(opt, git.GlobalScope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not unset Git config '%s'.", opt)
		ctx.Log.Info("Decision to delete LFS hooks is reset and now left to the user.")

	case opts.Print:
		conf := ctx.GitX.GetConfig(opt, git.GlobalScope)
		switch conf { //nolint:staticcheck
		case "y":
			fallthrough
		case "a" /* legacy */ :
			ctx.Log.Info("Detected LFS hooks are automatically deleted during install.")
		case "n":
			fallthrough
		case "s" /* legacy */ :
			ctx.Log.Info("Detected LFS hooks are not automatically deleted during install",
				"but instead backed up.")
		default:
			ctx.Log.Info("Deletion of detected LFS hooks is undefined and left to the user.")
		}

	default:
		cm.Panic("Wrong arguments.")
	}
}

func configListCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, gitOpts *GitOptions) {

	listCmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "Lists settings of the Githooks configuration.",
		Long: `Lists the Githooks related settings of the Githooks configuration.
Can be either global or local configuration, or both by default.`,
		PreRun: ccm.PanicIfAnyArgs(ctx.Log),
		Run: func(cmd *cobra.Command, args []string) {

			if !gitOpts.Local && !gitOpts.Global {
				_, _, _, err := ctx.GitX.GetRepoRoot()
				gitOpts.Local = err == nil
				gitOpts.Global = true

			} else if gitOpts.Local {
				ccm.AssertRepoRoot(ctx)
			}

			runList(ctx, gitOpts)
		}}

	listCmd.Flags().BoolVar(&gitOpts.Local, "local", false, "Use the local Git configuration.")
	listCmd.Flags().BoolVar(&gitOpts.Global, "global", false, "Use the global Git configuration.")
	configCmd.AddCommand(listCmd)
}

func configDisableCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions, gitOpts *GitOptions) {

	disableCmd := &cobra.Command{
		Use:   "disable [flags]",
		Short: "Disables Githooks in the current repository or globally.",
		Long: `Disables Githooks in the current repository or globally.
LFS hooks and replaced previous hooks are still executed by Githooks.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !gitOpts.Local && !gitOpts.Global {
				gitOpts.Local = true
			}

			runDisable(ctx, setOpts, gitOpts)
		}}

	optsPSR := createOptionMap(true, false, true)

	configSetOptions(disableCmd, setOpts, &optsPSR, ctx.Log, 0, 0)
	disableCmd.Flags().BoolVar(&gitOpts.Local, "local", false, "Use the local Git configuration (default).")
	disableCmd.Flags().BoolVar(&gitOpts.Global, "global", false, "Use the global Git configuration.")
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, disableCmd))
}

func configContainerizedHooksEnabledCmd(
	ctx *ccm.CmdContext,
	configCmd *cobra.Command,
	setOpts *SetOptions,
	gitOpts *GitOptions) {

	enableCmd := &cobra.Command{
		Use:   "enable-containerized-hooks [flags]",
		Short: "Enable running hooks containerized.",
		Long:  `Enable running hooks containerized over a container manager.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !gitOpts.Local && !gitOpts.Global {
				gitOpts.Local = true
			}

			runContainerizedHooksEnable(ctx, setOpts, gitOpts)
		}}

	optsPSR := createOptionMap(true, false, true)

	configSetOptions(enableCmd, setOpts, &optsPSR, ctx.Log, 0, 0)
	enableCmd.Flags().BoolVar(&gitOpts.Local, "local", false, "Use the local Git configuration (default).")
	enableCmd.Flags().BoolVar(&gitOpts.Global, "global", false, "Use the global Git configuration.")
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, enableCmd))
}

func configContainerManagerTypesCmd(
	ctx *ccm.CmdContext,
	configCmd *cobra.Command,
	setOpts *SetOptions,
	gitOpts *GitOptions) {

	enableCmd := &cobra.Command{
		Use:   "container-manager-types [flags]",
		Short: "Set container manger types to use (see 'enable-containerized-hooks').",
		Long: `Set container manager types to use where the first valid one is taken and used.
If unset 'docker' is used.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !gitOpts.Local && !gitOpts.Global {
				gitOpts.Local = true
			}

			runContainerManagerTypes(ctx, setOpts, gitOpts)
		}}

	optsPSR := createOptionMap(true, false, true)

	configSetOptions(enableCmd, setOpts, &optsPSR, ctx.Log, 1, 2) // nolint: mnd
	enableCmd.Flags().BoolVar(&gitOpts.Local, "local", false, "Use the local Git configuration (default).")
	enableCmd.Flags().BoolVar(&gitOpts.Global, "global", false, "Use the global Git configuration.")
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, enableCmd))
}

func configSearchDirCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions) {

	searchDirCmd := &cobra.Command{
		Use:   "search-dir [flags]",
		Short: "Changes the search directory used during installation.",
		Long: `Changes the previous search directory setting
used during installation.`,
		Run: func(cmd *cobra.Command, args []string) {
			runSearchDir(ctx, setOpts)
		}}

	optsPSR := createOptionMap(true, false, true)

	configSetOptions(searchDirCmd, setOpts, &optsPSR, ctx.Log, 1, 1)
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, searchDirCmd))
}

func configCloneURLCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions) {

	cloneURLCmd := &cobra.Command{
		Use:   "clone-url [flags]",
		Short: "Changes the Githooks clone url used for any update.",
		Long:  `Changes the Githooks clone url used for any update.`,
		Run: func(cmd *cobra.Command, args []string) {
			runCloneURL(ctx, setOpts)
		}}

	optsPrintSet := createOptionMap(true, false, false)

	configSetOptions(cloneURLCmd, setOpts, &optsPrintSet, ctx.Log, 1, 1)
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, cloneURLCmd))
}

func configCloneBranchCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions) {

	cloneBranchCmd := &cobra.Command{
		Use:   "clone-branch [flags]",
		Short: "Changes the Githooks clone url used for any update.",
		Long:  `Changes the Githooks clone url used for any update.`,
		Run: func(cmd *cobra.Command, args []string) {
			runCloneBranch(ctx, setOpts)
		}}

	optsPSR := createOptionMap(true, false, true)

	configSetOptions(cloneBranchCmd, setOpts, &optsPSR, ctx.Log, 1, 1)
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, cloneBranchCmd))
}

func configTrustAllHooksCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions) {

	trustCmd := &cobra.Command{
		Use:   "trust-all [flags]",
		Short: "Change trust settings in the current repository.",
		Long: `Change the trust setting in the current repository.

This command needs to be run at the root of a repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			runTrustAllHooks(ctx, setOpts)
		}}

	optsPSUR := createOptionMap(true, true, true)
	optsPSUR.Set = "accept"
	optsPSUR.SetDesc = "Accepts changes to all existing and new hooks\n" +
		"in the current repository when the trust marker\nis present."
	optsPSUR.Unset = "deny"
	optsPSUR.UnsetDesc = "Marks the repository as it has refused to\n" +
		"trust the changes, even if the trust marker is present."
	optsPSUR.ResetDesc = "Clears the trust setting."

	configSetOptions(trustCmd, setOpts, &optsPSUR, ctx.Log, 0, 0)
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, trustCmd))
}

func configUpdateCheckCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions) {

	updateCmd := &cobra.Command{
		Use:   "update-check [flags]",
		Short: "Change Githooks update-check settings.",
		Long:  `Enable or disable automatic Githooks update checks.`,
		Run: func(cmd *cobra.Command, args []string) {
			RunUpdateCheck(ctx, setOpts)
		}}

	optsPSUR := createOptionMap(true, true, true)
	wrapToEnableDisable(&optsPSUR)
	optsPSUR.SetDesc = "Enables automatic update checks for Githooks."
	optsPSUR.UnsetDesc = "Disables automatic update checks for Githooks."

	configSetOptions(updateCmd, setOpts, &optsPSUR, ctx.Log, 0, 0)
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, updateCmd))

}

func configUpdateTimeCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions) {

	updateTimeCmd := &cobra.Command{
		Use:   "update-time [flags]",
		Short: "Changes the Githooks update time.",
		Long: `Changes the Githooks update time used to check for updates.

Resets the last Githooks update time with the '--reset' option,
causing the update check to run next time if it is enabled.
Use 'git hooks update [--enable|--disable]' to change that setting.`,
		Run: func(cmd *cobra.Command, args []string) {
			runUpdateTime(ctx, setOpts)
		}}

	optsPR := createOptionMap(false, false, true)

	configSetOptions(updateTimeCmd, setOpts, &optsPR, ctx.Log, 0, 0)
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, updateTimeCmd))
}

func configSharedCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions, gitOpts *GitOptions) {

	sharedCmd := &cobra.Command{
		Use:   "shared [flags] [<git-url>...]",
		Short: "Updates the list of local or global shared hook repositories.",
		Long: `Updates the list of local or global shared hook repositories.

The '--add' option accepts multiple '<git-url>' arguments,
each containing a clone URL of a shared hook repository which gets added.`,
		Run: func(cmd *cobra.Command, args []string) {

			if !gitOpts.Local && !gitOpts.Global {
				_, _, _, err := ctx.GitX.GetRepoRoot()
				gitOpts.Global = true
				gitOpts.Local = setOpts.Print && err == nil
			} else if gitOpts.Local {
				ccm.AssertRepoRoot(ctx)
			}

			runSharedRepos(ctx, setOpts, gitOpts)
		}}

	optsPSR := createOptionMap(true, false, true)
	optsPSR.Set = "add"
	optsPSR.SetDesc = "Adds given shared hook repositories '<git-url>'s."
	sharedCmd.Flags().BoolVar(&gitOpts.Local, "local", false, "Use the local Git configuration.")
	sharedCmd.Flags().BoolVar(&gitOpts.Global, "global", false, "Use the global Git configuration (default).")

	configSetOptions(sharedCmd, setOpts, &optsPSR, ctx.Log, 1, -1)
	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, sharedCmd))
}

func configSkipNonExistingSharedHooks(
	ctx *ccm.CmdContext,
	configCmd *cobra.Command,
	setOpts *SetOptions,
	gitOpts *GitOptions) {

	nonExistSharedCmd := &cobra.Command{
		Use:   "skip-non-existing-shared-hooks [flags]",
		Short: "Enable or disable skipping non-existing shared hooks.",
		Long: `Enable or disable failing hooks with an error when any
shared hooks are missing. This usually means 'git hooks shared update'
has not been called yet.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !gitOpts.Local && !gitOpts.Global {
				gitOpts.Local = true
			}

			if gitOpts.Local {
				ccm.AssertRepoRoot(ctx)
			}

			runSkipNonExistingSharedHooks(ctx, setOpts, gitOpts)
		}}

	optsPSUR := createOptionMap(true, true, true)
	wrapToEnableDisable(&optsPSUR)
	optsPSUR.SetDesc = "Enable skipping non-existing shared hooks."
	optsPSUR.UnsetDesc = "Disable skipping non-existing shared hooks."
	optsPSUR.ResetDesc = "Reset skipping non-existing shared hooks."

	configSetOptions(nonExistSharedCmd, setOpts, &optsPSUR, ctx.Log, 0, 0)

	nonExistSharedCmd.Flags().BoolVar(&gitOpts.Local, "local", false,
		"Use the local Git configuration (default, except for '--print').")
	nonExistSharedCmd.Flags().BoolVar(&gitOpts.Global,
		"global", false, "Use the global Git configuration.")

	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, nonExistSharedCmd))
}

func configDisableSharedHooksUpdate(
	ctx *ccm.CmdContext,
	configCmd *cobra.Command,
	setOpts *SetOptions,
	gitOpts *GitOptions) {

	nonExistSharedCmd := &cobra.Command{
		Use:   "disable-shared-hooks-update [flags]",
		Short: "Disable/enable automatic updates of shared hooks.",
		Long: `Disable/enable automatic updates of shared hook
repositories.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !gitOpts.Local && !gitOpts.Global {
				gitOpts.Local = true
			}

			if gitOpts.Local {
				ccm.AssertRepoRoot(ctx)
			}

			runDisableSharedHooksUpdate(ctx, setOpts, gitOpts)
		}}

	optsPSUR := createOptionMap(true, true, true)
	optsPSUR.SetDesc = "Disable automatic updates of shared hooks."
	optsPSUR.UnsetDesc = "Enabled automatic updates of shared hooks.."
	optsPSUR.ResetDesc = "Reset (enable) automatic updates of shared hooks."

	configSetOptions(nonExistSharedCmd, setOpts, &optsPSUR, ctx.Log, 0, 0)

	nonExistSharedCmd.Flags().BoolVar(&gitOpts.Local, "local", false,
		"Use the local Git configuration (default, except for '--print').")
	nonExistSharedCmd.Flags().BoolVar(&gitOpts.Global,
		"global", false, "Use the global Git configuration.")

	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, nonExistSharedCmd))
}

func configFailUntrustedHooks(
	ctx *ccm.CmdContext,
	configCmd *cobra.Command,
	setOpts *SetOptions,
	gitOpts *GitOptions) {

	nonExistSharedCmd := &cobra.Command{
		Use:   "skip-untrusted-hooks [flags]",
		Short: "Enable/disable skipping active, untrusted hooks.",
		Long: `Enable or disable failing hooks with an error when any
active, untrusted hooks are present.
Mostly wanted if all hooks must be executed.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !gitOpts.Local && !gitOpts.Global {
				gitOpts.Local = true
			}

			if gitOpts.Local {
				ccm.AssertRepoRoot(ctx)
			}

			runSkipUntrustedHooks(ctx, setOpts, gitOpts)
		}}

	optsPSUR := createOptionMap(true, true, true)
	wrapToEnableDisable(&optsPSUR)
	optsPSUR.SetDesc = "Enable skipping active, untrusted hooks."
	optsPSUR.UnsetDesc = "Disable skipping active, untrusted hooks."
	optsPSUR.ResetDesc = "Reset skipping active, untrusted hooks."

	configSetOptions(nonExistSharedCmd, setOpts, &optsPSUR, ctx.Log, 0, 0)

	nonExistSharedCmd.Flags().BoolVar(&gitOpts.Local, "local", false,
		"Use the local Git configuration (default, except for '--print').")
	nonExistSharedCmd.Flags().BoolVar(&gitOpts.Global,
		"global", false, "Use the global Git configuration.")

	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, nonExistSharedCmd))
}

func configNonInteractiveRunner(
	ctx *ccm.CmdContext,
	configCmd *cobra.Command,
	setOpts *SetOptions,
	gitOpts *GitOptions) {

	nonInteracticeRunner := &cobra.Command{
		Use:   "non-interactive-runner [flags]",
		Short: "Enables/disables non-interactive execution of the runner.",
		Long: `Enable or disables non-interactive execution of
the Githooks runner executable.

Enabling non-interactivity will only default answer all non-fatal prompts.
Fatal prompts (e.g. the trust prompts) still need to be configured to pass.
See 'git hooks config trust-all --help'.`,
		Run: func(cmd *cobra.Command, args []string) {
			if gitOpts.Local {
				ccm.AssertRepoRoot(ctx)
			}

			runRunnerNonInteractive(ctx, setOpts, gitOpts)
		}}

	optsPSUR := createOptionMap(true, true, true)
	wrapToEnableDisable(&optsPSUR)
	optsPSUR.SetDesc = "Enables non-interactive mode of the runner executable."
	optsPSUR.UnsetDesc = "Disables non-interactive mode of the runner executable."
	optsPSUR.ResetDesc = "Reset non-interactive mode of the runner executable."

	configSetOptions(nonInteracticeRunner, setOpts, &optsPSUR, ctx.Log, 0, 0)

	nonInteracticeRunner.Flags().BoolVar(&gitOpts.Local, "local", false,
		"Use the local Git configuration (default, except for '--print').")
	nonInteracticeRunner.Flags().BoolVar(&gitOpts.Global, "global", false,
		"Use the global Git configuration.")

	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, nonInteracticeRunner))
}

func configDetectedLFSCmd(ctx *ccm.CmdContext, configCmd *cobra.Command, setOpts *SetOptions, gitOpts *GitOptions) {

	deleteDetectedLFSCmd := &cobra.Command{
		Use:   "delete-detected-lfs-hooks [flags]",
		Short: "Change the behavior for detected LFS hooks during install.",
		Long: `By default, detected LFS hooks during install are
disabled and backed up.`,
		Run: func(cmd *cobra.Command, args []string) {
			runDeleteDetectedLFSHooks(ctx, setOpts)
		}}

	optsPSUR := createOptionMap(true, true, true)
	wrapToEnableDisable(&optsPSUR)

	optsPSUR.SetDesc = "Remember to always delete detected LFS hooks\n" +
		"instead of the default behavior."
	optsPSUR.UnsetDesc = "Remember to always not delete detected LFS hooks and\n" +
		"to resort to the default behavior."
	optsPSUR.ResetDesc = "Resets the decision to the default behavior."

	configSetOptions(deleteDetectedLFSCmd, setOpts, &optsPSUR, ctx.Log, 0, 0)

	configCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, deleteDetectedLFSCmd))
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	configCmd := &cobra.Command{
		Use:    "config",
		Short:  "Manages various Githooks configuration.",
		Long:   `Manages various Githooks configuration.`,
		PreRun: ccm.PanicWrongArgs(ctx.Log)}

	gitOpts := GitOptions{}
	setOpts := SetOptions{}

	configListCmd(ctx, configCmd, &gitOpts)
	configDisableCmd(ctx, configCmd, &setOpts, &gitOpts)
	configTrustAllHooksCmd(ctx, configCmd, &setOpts)

	configSearchDirCmd(ctx, configCmd, &setOpts)
	configUpdateCheckCmd(ctx, configCmd, &setOpts)
	configUpdateTimeCmd(ctx, configCmd, &setOpts)
	configCloneURLCmd(ctx, configCmd, &setOpts)
	configCloneBranchCmd(ctx, configCmd, &setOpts)

	configContainerizedHooksEnabledCmd(ctx, configCmd, &setOpts, &gitOpts)
	configContainerManagerTypesCmd(ctx, configCmd, &setOpts, &gitOpts)

	configSharedCmd(ctx, configCmd, &setOpts, &gitOpts)
	configDisableSharedHooksUpdate(ctx, configCmd, &setOpts, &gitOpts)

	configSkipNonExistingSharedHooks(ctx, configCmd, &setOpts, &gitOpts)
	configFailUntrustedHooks(ctx, configCmd, &setOpts, &gitOpts)

	configNonInteractiveRunner(ctx, configCmd, &setOpts, &gitOpts)

	configDetectedLFSCmd(ctx, configCmd, &setOpts, &gitOpts)

	configCmd.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		ccm.CheckGithooksSetup(ctx.Log, ctx.GitX)
	}

	return ccm.SetCommandDefaults(ctx.Log, configCmd)
}
