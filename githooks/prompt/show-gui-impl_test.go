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
	promptCtx, _ := CreateContext(log, &cm.ExecContext{}, nil, true, false)

	ans, err := promptCtx.ShowPrompt("Enter a string sssssssssss:", "This is the default string", ValidatorAnswerNotEmpty)
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	ans, err = promptCtx.ShowPromptOptions("Choose string      ss       ssssssssssssss s s s asd asdfl kjj sdlfök jsaölkdf jölaskdjf lökasjd flökjsa döfl  s:", "(Yes/no)", "Y/n", "Yes", "No")
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	ans, err = promptCtx.ShowPromptOptions(
		"This string sssssssssss s s as sd asd asd asd\nasd asd asd s              asd asd asd asd asd asd?", "(Yes/no/skip/skip all)", "Y/n/s/a", "Yes", "No", "Skip", "Skip All")
	log.InfoF("Answer: '%s'", ans)
	log.AssertNoErrorF(err, "Error occurred.")

	a, e := promptCtx.ShowPromptMulti("Enter strings ('exit' cancels):", "exit", ValidatorAnswerNotEmpty)
	log.InfoF("Answer: '%+q'", a)
	log.AssertNoErrorF(e, "Error occurred.")
}
