package prompt

import (
	"bufio"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"io"
	"os"
	"runtime"
	"strings"
)

// AnswerValidator is the callback type for the answer validator.
type AnswerValidator func(string) error

// IContext defines the interface to show a prompt to the user.
type IContext interface {
	ShowPromptOptions(
		text string,
		hintText string,
		shortOptions string,
		longOptions ...string) (string, error)

	ShowPrompt(
		text string,
		defaultAnswer string,
		validator AnswerValidator) (string, error)

	ShowPromptMulti(
		text string,
		exitAnswer string,
		validator AnswerValidator) ([]string, error)

	Close()
}

// Formatter is the format function to format a prompt.
type Formatter func(format string, args ...interface{}) string

// Context defines the prompt context based on a `ILogContext`
// or as a fallback using the defined dialog tool if configured.
type Context struct {
	log cm.ILogContext

	useGUI bool

	// Terminal data
	promptFmt     Formatter
	errorFmt      Formatter
	termOut       io.Writer
	termIn        *os.File
	termInScanner *bufio.Scanner

	// Promp settings
	printAnswer     bool
	maxTries        uint
	panicIfMaxTries bool

	// Prompt over the tool script
	// if existing.
	execCtx cm.IExecContext
	tool    cm.IExecutable
}

// Close closes the prompt context.
func (p *Context) Close() {
	if p.termIn != nil {
		p.termIn.Close()
	}
}

// CreateContext creates a prompt context `IContext`.
// The GUI dialog gets only used if no terminal is attached on the output.
func CreateContext(
	log cm.ILogContext,
	execCtx cm.IExecContext,
	tool cm.IExecutable,
	useGUIFallback,
	useStdIn bool) (IContext, error) {

	var err error

	var input *os.File
	printAnswer := false
	maxTries := uint(3) //nolint: gomnd

	if useStdIn {
		input = os.Stdin
		printAnswer = true
		maxTries = uint(1) //nolint: gomnd
	} else {
		input, err = cm.GetCtty()
		// if err != nil => we don't have a terminal attached.
	}

	var output io.Writer
	if useStdIn || (!AssertOutputIsTerminal || log.IsInfoATerminal()) {
		// We use the output of the log:
		// - If we are using stdin it does not matter if the output is really a terminal.
		// - Otherwise its crucial that it is a terminal, since the prompt can not
		//   be shown and the user has not notion what to input, in that case
		//   output == nil, which results in the default answer.
		// For tests: AssertOutputIsTerminal == false, which always sets the output.
		output = log.GetInfoWriter()
	}

	p := Context{
		log: log,

		// In case we have no terminal input, and no output to show the prompt
		// fallback to using the GUI if enabled.
		useGUI: EnableGUI && useGUIFallback && (input == nil || output == nil),

		errorFmt:      log.GetErrorFormatter(true),
		promptFmt:     log.GetPromptFormatter(true),
		termOut:       output,
		termIn:        input,
		termInScanner: bufio.NewScanner(input),

		maxTries:        maxTries,
		panicIfMaxTries: true,
		printAnswer:     printAnswer,

		execCtx: execCtx,
		tool:    tool}

	runtime.SetFinalizer(&p, func(p *Context) { p.Close() })

	return &p, err
}

func getDefaultAnswer(options []string) (string, int) {
	for idx, r := range options {
		if strings.ToLower(r) != r { // is it an upper case letter?
			return strings.ToLower(r), idx
		}
	}

	return "", -1
}

// CreateValidatorAnswerOptions creates a validator which validates against
// a list of options.
func CreateValidatorAnswerOptions(options []string) AnswerValidator {

	return func(answer string) error {

		correct := strs.Any(
			options,
			func(o string) bool {
				return strings.EqualFold(answer, o)
			})

		if !correct {
			return cm.ErrorF("Answer '%s' not in '%q'.", answer, options)
		}

		return nil
	}
}

// ValidatorAnswerNotEmpty checks that answers are non-empty.
var ValidatorAnswerNotEmpty AnswerValidator = func(s string) error {
	if strs.IsEmpty(strings.TrimSpace(s)) {
		return cm.Error("Answer must not be empty.")
	}

	return nil
}

// CreateValidatorIsDirectory creates a answer validator
// which checks existing paths.
func CreateValidatorIsDirectory(tildeRepl string) AnswerValidator {
	return func(s string) error {
		s = cm.ReplaceTildeWith(s, tildeRepl)
		if !cm.IsDirectory(s) {
			return cm.Error("Answer must be an existing directory.")
		}

		return nil
	}
}
