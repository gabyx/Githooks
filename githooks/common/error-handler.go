package common

import (
	"os"
	"strings"
)

// HandleCLIErrors generally handles errors for the Githooks executables. Argument `cwd` can be empty.
func HandleCLIErrors(
	err interface{},
	cwd string,
	log ILogContext,
	getBugReportingInfo func(string) (string, error)) bool {

	if err == nil {
		return false
	}

	var message []string
	withTrace := false

	switch v := err.(type) {
	case GithooksFailure:
		message = append(message, "Fatal error -> Abort.")
	case error:
		info, e := getBugReportingInfo(cwd)
		v = CombineErrors(v, e)
		message = append(message, v.Error(), info)
		withTrace = true

	default:
		info, e := getBugReportingInfo(cwd)
		e = CombineErrors(Error("Panic 💩: Unknown error."), e)
		message = append(message, e.Error(), info)
		withTrace = true
	}

	if log != nil {
		if withTrace {
			log.ErrorWithStacktrace(message...)
		} else {
			log.Error(message...)
		}
	} else {
		os.Stderr.WriteString(strings.Join(message, "\n"))
	}

	return true
}
