// +build linux windows darwin

package gui

import (
	"strings"

	cm "gabyx/githooks/common"
	pcm "gabyx/githooks/prompt/common"

	"github.com/gen2brain/dlgs"
)

func ShowPromptDialog(
	title string,
	text string,
	defaultAnswer string) (string, error) {

	ans, success, err := dlgs.Entry(title, text, defaultAnswer)

	if err != nil {
		return "", cm.CombineErrors(err, cm.Error("GUI prompt dialog failed."))
	} else if !success {
		return "", pcm.CancledError
	}

	return ans, nil
}

func ShowOptionsDialog(
	title string,
	question string,
	defaultAnswerIdx int,
	options []string,
	longOptions []string) (string, error) {

	if len(options) == 2 { // nolint: gomnd

		o1 := strings.ToLower(options[0])
		o2 := strings.ToLower(options[1])

		if (o1 == "y" && o2 == "n") ||
			(o1 == "n" && o2 == "y") {
			// This is a normal yes/no prompt

			defaultCancel := (defaultAnswerIdx == 0 && o1 == "n") ||
				(defaultAnswerIdx == 1 && o2 == "n")

			yes, err := dlgs.Question(title, question, defaultCancel)
			if err != nil {
				return "",
					cm.CombineErrors(err, cm.Error("GUI option dialog failed."))
			}

			if yes {
				return "y", nil
			}

			return "n", nil
		}
	}

	// Make a list dialog
	selected, success, err := dlgs.List(title, question, longOptions)
	if err != nil {
		return "", cm.CombineErrors(err, cm.Error("GUI option dialog failed."))
	} else if !success {
		return "", pcm.CancledError
	}

	// Get the chosen answer
	for i := range longOptions {
		if selected == longOptions[i] {
			return strings.ToLower(options[i]), nil
		}
	}

	// User has not chosen anything or canceled...
	return "", nil
}
