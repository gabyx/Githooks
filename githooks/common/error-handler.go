package common

import (
	"os"
	"strings"
)

// HandleCLIErrors generally handles errors for the Githooks executables. Argument `cwd` can be empty.
func HandleCLIErrors(
	err any,
	log ILogContext,
	getBugReportingInfo func() string,
) bool {

	if err == nil {
		return false
	}

	var message []string
	withTrace := false

	switch v := err.(type) {
	case GithooksFailure:
		message = append(message, "Fatal error -> Abort.")
	case error:
		message = append(message, v.Error(), getBugReportingInfo())
		withTrace = true

	default:
		message = append(message, "Panic 💩: Unknown error.", getBugReportingInfo())
		withTrace = true
	}

	if log != nil {
		if withTrace {
			log.ErrorWithStacktrace(message...)
		} else {
			log.Error(message...)
		}
	} else {
		_, _ = os.Stderr.WriteString(strings.Join(message, "\n"))
		_, _ = os.Stderr.WriteString("\n")
	}

	return true
}
