//go:build dontbuild

package prompt_test

import (
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/prompt"
	"os"

	"testing"
)

func TestCoverage(t *testing.T) {

	log, err := cm.CreateLogContext(false)
	cm.AssertNoErrorPanic(err)

	os.Stdin = nil
	promptCtx, _ := prompt.CreateContext(log, prompt.ToolContext{}, false, false)

	ans, err := promptCtx.ShowEntry("Enter a default string:",
		"This is the default string", prompt.ValidatorAnswerNotEmpty)
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	ans, err = promptCtx.ShowOptions("Do you wanna do it?", "(Yes/no)", "Y/n", "Yes", "No")
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	ans, err = promptCtx.ShowOptions(
		"Do you really wanna do it because its gonna get really messy and output will be convoluted?",
		"(Yes/no/skip/skip all)", "Y/n/s/a", "Yes", "No", "Skip", "Skip All")
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	a, e := promptCtx.ShowEntryMulti("Enter strings", "exit",
		prompt.ValidatorAnswerNotEmpty)
	log.InfoF("Answer: '%+q'", a)
	log.AssertNoErrorF(e, "Error occurred.")
}
