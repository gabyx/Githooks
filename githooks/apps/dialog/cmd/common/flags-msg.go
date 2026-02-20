package common

import (
	"github.com/gabyx/githooks/githooks/apps/dialog/settings"
	cm "github.com/gabyx/githooks/githooks/common"
)

type iconArgs struct {
	icon *settings.DialogIcon
}

func (i *iconArgs) String() string {
	return ""
}

func (i *iconArgs) Type() string {
	return "DialogIcon"
}

func (i *iconArgs) Set(s string) error {
	switch s {
	case "info":
		*i.icon = settings.InfoIcon
	case "warning":
		*i.icon = settings.WarningIcon
	case "error":
		*i.icon = settings.ErrorIcon
	case "question":
		*i.icon = settings.QuestionIcon
	default:
		return cm.ErrorF(
			"Icon name '%s' is not one of ['info', 'warning', 'error', 'question'].",
			s,
		)
	}

	return nil
}

type msgStyleArgs struct {
	style *settings.MessageStyle
}

func (i *msgStyleArgs) String() string {
	return ""
}

func (i *msgStyleArgs) Type() string {
	return "MessageStyle"
}

func (i *msgStyleArgs) Set(s string) error {
	switch s {
	case "info":
		*i.style = settings.InfoStyle
	case "warning":
		*i.style = settings.WarningStyle
	case "error":
		*i.style = settings.ErrorStyle
	case "question":
		*i.style = settings.QuestionStyle
	default:
		return cm.ErrorF(
			"Style name '%s' is not one of ['info', 'warning', 'error', 'question'].",
			s,
		)
	}

	return nil
}
