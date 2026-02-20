//go:build dontbuild

package prompt_test

import (
	"os"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/prompt"

	"testing"
)

func TestCoverage(t *testing.T) {

	log, err := cm.CreateLogContext(false, false)
	cm.AssertNoErrorPanic(err)

	os.Stdin = nil
	promptx, _ := prompt.CreateContext(log, false, false)

	ans, err := promptx.ShowEntry("Enter a default string:",
		"This is the default string", prompt.ValidatorAnswerNotEmpty)
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	ans, err = promptx.ShowOptions("Do you wanna do it?", "(Yes/no)", "Y/n", "Yes", "No")
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	ans, err = promptx.ShowOptions(
		"Do you really wanna do it because its gonna get really messy and output will be convoluted?",
		"(Yes/no/skip/skip all)",
		"Y/n/s/a",
		"Yes",
		"No",
		"Skip",
		"Skip All",
	)
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	a, e := promptx.ShowEntryMulti("Enter strings", "exit",
		prompt.ValidatorAnswerNotEmpty)
	log.InfoF("Answer: '%+q'", a)
	log.AssertNoErrorF(e, "Error occurred.")

	e = promptx.ShowMessage("This is a warning prompt message", true)
	log.AssertNoErrorF(e, "Error occurred.")
}
