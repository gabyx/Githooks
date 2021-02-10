// +build gui

package prompt

import (
	cm "gabyx/githooks/common"
	"os"

	"testing"
)

func TestCoverage(t *testing.T) {

	log, err := cm.CreateLogContext(false)
	cm.AssertNoErrorPanic(err)

	os.Stdin = nil
	promptCtx, err := CreateContext(log, &cm.ExecContext{}, nil, true, true)
	log.AssertNoErrorPanic(err)

	ans, err := promptCtx.ShowPrompt("Enter a string:", "This is the default string", ValidatorAnswerNotEmpty)
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occured.")

	ans, err = promptCtx.ShowPromptOptions("Enter a string:", "(Yes/no)", "Y/n", "Yes", "No")
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occured.")

	ans, err = promptCtx.ShowPromptOptions("Enter a string:", "(Yes/no/skip/skip all)", "Y/n/s/a", "Yes", "No", "Skip", "Skip All")
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occured.")

	a, e := promptCtx.ShowPromptMulti("Enter a strings:", "exit", ValidatorAnswerNotEmpty)
	log.InfoF("Answer: '%+q'", a)
	log.AssertNoErrorF(e, "Error occured.")
}
