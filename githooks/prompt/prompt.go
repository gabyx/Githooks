package prompt

import (
	"bufio"
	"io"
	"os"
	"runtime"

	cm "github.com/gabyx/githooks/githooks/common"
)

// AnswerValidator is the callback type for the answer validator.
type AnswerValidator func(string) error

// IContext defines the interface to show a prompt to the user.
type IContext interface {
	ShowOptions(
		text string,
		hintText string,
		shortOptions string,
		longOptions ...string) (string, error)

	ShowEntry(
		text string,
		defaultAnswer string,
		validator AnswerValidator) (string, error)

	ShowEntryMulti(
		text string,
		exitAnswer string,
		validator AnswerValidator) ([]string, error)

	Close()
}

// Formatter is the format function to format a prompt or error.
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
	termIn        io.Reader
	termInScanner *bufio.Scanner

	// Promp settings
	printAnswer     bool
	maxTries        uint
	panicIfMaxTries bool
}

// Close closes the prompt context.
func (p *Context) Close() {
	if p.termIn != nil {
		t, ok := p.termIn.(*os.File)
		if ok {
			t.Close()
		}
	}
}

// CreateContext creates a prompt context `IContext`.
// The GUI dialog gets only used if no terminal is attached on the output.
func CreateContext(
	log cm.ILogContext,
	useGUIFallback,
	useStdIn bool) (IContext, error) {

	var err error

	var input io.Reader
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
		printAnswer:     printAnswer}

	runtime.SetFinalizer(&p, func(p *Context) { p.Close() })

	return &p, err
}

// Context for the dialog tool script/executable, if one is installed.
// If `execCtx` and `tool` is nil, the context is not setup and will not be used.
type ToolContext struct {
	execCtx cm.IExecContext
	tool    cm.IExecutable
}

// IsExecutable tells if the prompt context for the dialog tool is executable.
func (p *ToolContext) IsSetup() bool {
	return p.execCtx != nil && p.tool != nil
}

// Creates a prompt context for the dialog tool script/executable.
func CreateToolContext(execCtx cm.IExecContext, tool cm.IExecutable) (ToolContext, error) {
	return ToolContext{execCtx: execCtx, tool: tool}, nil
}
