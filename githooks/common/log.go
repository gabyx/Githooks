package common

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/gookit/color"
	"golang.org/x/term"
)

const (
	// GithooksEmoji is the general Githooks emoji.
	GithooksEmoji = "🦎"

	githooksSuffix    = "" // If you like you can make it: "Githooks: "
	debugSuffix       = "🛠  " + githooksSuffix
	infoSuffix        = GithooksEmoji + " " + githooksSuffix
	warnSuffix        = "⛑  " + githooksSuffix
	errorSuffix       = "⛔ "
	promptSuffix      = "❓ " + githooksSuffix
	informationSuffix = "ℹ️   "
	indent            = "   "

	// ListItemLiteral is the list item used for CLI and other printing stuff.
	ListItemLiteral = "•"
)

var colorInfo = color.FgBlue.Render
var colorError = color.FgRed.Render
var colorPrompt = color.FgGreen.Render

// ILogContext defines the log interface.
type ILogContext interface {
	// Log functions
	Debug(lines ...string)
	DebugF(format string, args ...any)
	Info(lines ...string)
	InfoF(format string, args ...any)
	Warn(lines ...string)
	WarnF(format string, args ...any)
	Error(lines ...string)
	ErrorF(format string, args ...any)
	ErrorWithStacktrace(lines ...string)
	ErrorWithStacktraceF(format string, args ...any)
	Panic(lines ...string)
	PanicF(format string, args ...any)

	// Assert helper functions
	ErrorOrPanicF(isFatal bool, err error, format string, args ...any)
	ErrorOrPanicIfF(isFatal bool, condition bool, format string, args ...any)
	AssertWarn(condition bool, lines ...string)
	AssertWarnF(condition bool, format string, args ...any)
	DebugIf(condition bool, lines ...string)
	DebugIfF(condition bool, format string, args ...any)
	InfoIf(condition bool, lines ...string)
	InfoIfF(condition bool, format string, args ...any)
	WarnIf(condition bool, lines ...string)
	WarnIfF(condition bool, format string, args ...any)
	ErrorIf(condition bool, lines ...string)
	ErrorIfF(condition bool, format string, args ...any)
	PanicIf(condition bool, lines ...string)
	PanicIfF(condition bool, format string, args ...any)
	AssertNoError(err error, lines ...string) bool
	AssertNoErrorF(err error, format string, args ...any) bool
	AssertNoErrorPanic(err error, lines ...string)
	AssertNoErrorPanicF(err error, format string, args ...any)

	HasColor() bool
	GetIndent() string

	GetInfoWriter() io.Writer
	GetInfoWriterOriginal() io.Writer
	IsInfoATerminal() bool

	GetErrorWriter() io.Writer
	IsErrorATerminal() bool

	AddFileWriter(file *os.File)
	GetFileWriter() *os.File
	MoveFileWriterToEnd()
	RemoveFileWriter()
}

// ILogStats is an interface for log statistics.
type ILogStats interface {
	ErrorCount() int
	WarningCount() int

	ResetStats()

	EnableStats()
	DisableStats()
}

type FormattedWriter struct {
	format func(...any) string
	writer io.Writer
}

func (w *FormattedWriter) Write(p []byte) (n int, err error) {
	s := []byte(w.format(string(p)))
	sn, err := w.writer.Write(s)
	if err != nil {
		return
	}
	if sn != len(s) {
		return sn, io.ErrShortWrite
	}

	// return always the input length, otherwise
	// writing fails to this Writer.
	return len(p), err
}

// Loglevel constants for runtime switching.
const (
	debugLevel = -1
	infoLevel  = 0
	warnLevel  = 1
	errorLevel = 2

	disableLevel = 10
)

// LogContext defines the data for a log context.
type LogContext struct {
	stdout *os.File
	stderr *os.File
	file   *os.File

	debug io.Writer
	info  io.Writer
	warn  io.Writer
	error io.Writer

	infoIsTerminal   bool
	errorIsTerminal  bool
	isColorSupported bool

	doTrackStats bool
	nWarnings    int
	nErrors      int

	level int
}

// NewColoredPromptWriter returns a colored prompt writer.
func NewColoredPromptWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return nil
	}

	return &FormattedWriter{format: colorPrompt, writer: writer}
}

// NewColoredInfoWriter returns a colored info writer.
func NewColoredInfoWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return nil
	}

	return &FormattedWriter{format: colorInfo, writer: writer}
}

// NewColoredErrorWriter returns a colored error writer.
func NewColoredErrorWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return nil
	}

	return &FormattedWriter{format: colorError, writer: writer}
}

// CreateLogContext creates a log context and if `setFromEnv` is set will
// set the log level from the environment variable `GITHOOKS_LOG_LEVEL` which can be
// `DebugLevel`, `InfoLevel`, `WarnLevel`, `ErrorLevel`, `DisableLevel`.
func CreateLogContext(onlyStderr bool, setLogLevelFromEnv bool) (*LogContext, error) {
	var l LogContext
	l.stdout = os.Stdout
	l.stderr = os.Stderr

	if onlyStderr {
		l.stdout = l.stderr
	}

	l.infoIsTerminal = term.IsTerminal(int(l.stdout.Fd()))
	l.errorIsTerminal = term.IsTerminal(int(l.stderr.Fd()))
	l.isColorSupported = (l.infoIsTerminal && l.errorIsTerminal) && color.IsSupportColor()

	l.setupWriters()
	l.setupLogLevel(setLogLevelFromEnv)

	l.doTrackStats = true

	return &l, nil
}

func (c *LogContext) setupWriters() {
	if c.HasColor() {
		c.debug = NewColoredInfoWriter(c.stdout)
		c.info = NewColoredInfoWriter(c.stdout)
		c.warn = NewColoredErrorWriter(c.stderr)
		c.error = NewColoredErrorWriter(c.stderr)
	} else {
		c.debug = c.stdout
		c.info = c.stdout
		c.warn = c.stderr
		c.error = c.stderr
	}
}

func parseLogLevel() int {
	level := strings.TrimSpace(os.Getenv("GITHOOKS_LOG_LEVEL"))

	switch level {
	case "debug":
		return debugLevel
	default: // nolint: gocritic
		fallthrough
	case "info":
		return infoLevel
	case "warn":
		return warnLevel
	case "error":
		return errorLevel
	case "disable":
		return disableLevel
	}
}

func (c *LogContext) setupLogLevel(fromEnv bool) {
	if DebugLog {
		// Always overwrites.
		c.level = debugLevel
	} else if fromEnv {
		c.level = parseLogLevel()
	}
}

// GetIndent returns the used indent.
func (c *LogContext) GetIndent() string {
	return indent
}

// HasColor returns if the log uses colors.
func (c *LogContext) HasColor() bool {
	return c.isColorSupported
}

// GetInfoWriter returns the info writer.
func (c *LogContext) GetInfoWriter() io.Writer {
	return c.info
}

// GetInfoWriter returns the original info writer.
func (c *LogContext) GetInfoWriterOriginal() io.Writer {
	return c.stdout
}

// GetErrorWriter returns the error writer.
func (c *LogContext) GetErrorWriter() io.Writer {
	return c.error
}

// IsInfoATerminal returns `true` if the info log is connected to a terminal.
func (c *LogContext) IsInfoATerminal() bool {
	return c.infoIsTerminal
}

// IsErrorATerminal returns `true` if the error log is connected to a terminal.
func (c *LogContext) IsErrorATerminal() bool {
	return c.errorIsTerminal
}

// Debug logs a debug message.
func (c *LogContext) Debug(lines ...string) {
	if DebugLog && c.level <= debugLevel {
		_, _ = fmt.Fprint(c.debug, FormatMessage(debugSuffix, indent, lines...), "\n")
	}
}

// DebugF logs a debug message.
func (c *LogContext) DebugF(format string, args ...any) {
	if DebugLog && c.level <= debugLevel {
		_, _ = fmt.Fprint(c.debug, FormatMessageF(debugSuffix, indent, format, args...), "\n")
	}
}

// Info logs a info message.
func (c *LogContext) Info(lines ...string) {
	if c.level <= infoLevel {
		_, _ = fmt.Fprint(c.info, FormatInfo(lines...), "\n")
	}
}

// InfoF logs a info message.
func (c *LogContext) InfoF(format string, args ...any) {
	if c.level <= infoLevel {
		_, _ = fmt.Fprint(c.info, FormatInfoF(format, args...), "\n")
	}
}

// Warn logs a warning message.
func (c *LogContext) Warn(lines ...string) {
	if c.level <= warnLevel {
		_, _ = fmt.Fprint(c.warn, FormatMessage(warnSuffix, indent, lines...), "\n")
	}
	if c.doTrackStats {
		c.nWarnings++
	}
}

// WarnF logs a warning message.
func (c *LogContext) WarnF(format string, args ...any) {
	if c.level <= warnLevel {
		_, _ = fmt.Fprint(c.warn, FormatMessageF(warnSuffix, indent, format, args...), "\n")
	}
	if c.doTrackStats {
		c.nWarnings++
	}
}

// Error logs an error.
func (c *LogContext) Error(lines ...string) {
	if c.level <= errorLevel {
		_, _ = fmt.Fprint(c.error, FormatMessage(errorSuffix, indent, lines...), "\n")
	}
	if c.doTrackStats {
		c.nErrors++
	}
}

// ErrorF logs an error.
func (c *LogContext) ErrorF(format string, args ...any) {
	if c.level <= errorLevel {
		_, _ = fmt.Fprint(c.error, FormatMessageF(errorSuffix, indent, format, args...), "\n")
	}
	if c.doTrackStats {
		c.nErrors++
	}
}

// FormatInfoMessage formats a info message.
func FormatInfoMessage(format string, args ...any) string {
	return FormatMessageF(infoSuffix, indent, format, args...)
}

// FormatInformationMessage formats a informational message.
func FormatInformationMessage(format string, args ...any) string {
	return FormatMessageF(informationSuffix, indent, format, args...)
}

// FormatErrorMessage formats an error message.
func FormatErrorMessage(format string, args ...any) string {
	return FormatMessageF(errorSuffix, indent, format, args...)
}

// FormatPromptMessage formats a prompt message.
func FormatPromptMessage(format string, args ...any) string {
	return FormatMessageF(promptSuffix, indent, format, args...)
}

// ErrorWithStacktrace logs and error with the stack trace.
func (c *LogContext) ErrorWithStacktrace(lines ...string) {
	stackLines := strs.SplitLines(string(debug.Stack()))
	lines = append(lines, "", "Stacktrace:", "-----------")
	c.Error(append(lines, stackLines...)...)
}

// ErrorWithStacktraceF logs and error with the stack trace.
func (c *LogContext) ErrorWithStacktraceF(format string, args ...any) {
	c.ErrorWithStacktrace(strs.Fmt(format, args...))
}

// Panic logs an error and calls panic with a GithooksFailure.
func (c *LogContext) Panic(lines ...string) {
	m := FormatMessage(errorSuffix, indent, lines...)

	if c.level <= errorLevel {
		_, _ = fmt.Fprint(c.error, m, "\n")
	}

	panic(GithooksFailure{m})
}

// PanicF logs an error and calls panic with a GithooksFailure.
func (c *LogContext) PanicF(format string, args ...any) {
	m := FormatMessageF(errorSuffix, indent, format, args...)

	if c.level <= errorLevel {
		_, _ = fmt.Fprint(c.error, m, "\n")
	}

	panic(GithooksFailure{m})
}

// WarningCount gets the number of logged warnings.
func (c *LogContext) WarningCount() int {
	return c.nWarnings
}

// ErrorCount gets the number of logged errors.
func (c *LogContext) ErrorCount() int {
	return c.nErrors
}

// ResetStats resets the log statistics.
func (c *LogContext) ResetStats() {
	c.nErrors = 0
	c.nWarnings = 0
}

// DisableStats disables the log statistics.
func (c *LogContext) DisableStats() {
	c.doTrackStats = false
}

// EnableStats enables the log statistics.
func (c *LogContext) EnableStats() {
	c.doTrackStats = false
}

// GetFileWriter gets a optional file writer.
func (c *LogContext) GetFileWriter() *os.File {
	return c.file
}

// AddFileWriter adds a another sink to all sinks log.
func (c *LogContext) AddFileWriter(file *os.File) {
	if file == nil {
		return
	}

	c.file = file

	if c.debug != nil && c.level <= debugLevel {
		c.debug = io.MultiWriter(c.debug, file)
	}

	if c.info != nil && c.level <= infoLevel {
		c.info = io.MultiWriter(c.info, file)
	}

	if c.warn != nil && c.level <= warnLevel {
		c.warn = io.MultiWriter(c.warn, file)
	}

	if c.error != nil && c.level <= errorLevel {
		c.error = io.MultiWriter(c.error, file)
	}
}

// RemoveFileWriter a potentially added file writer.
func (c *LogContext) RemoveFileWriter() {
	if c.file != nil {
		c.setupWriters()
		_ = c.file.Close()
	}

	c.file = nil
}

// MoveFileWriterToEnd the the write pointer to the end of the file.
func (c *LogContext) MoveFileWriterToEnd() {
	if c.file != nil {
		_, _ = c.file.Seek(0, 2) // nolint: mnd
	}
}

// FormatMessage formats  several lines with a suffix and indent.
func FormatMessage(suffix string, indent string, lines ...string) string {
	return suffix + strings.Join(lines, "\n"+indent)
}

// FormatMessageF formats  several lines with a suffix and indent.
func FormatMessageF(suffix string, indent string, format string, args ...any) string {
	s := suffix + strs.Fmt(format, args...)
	return strings.ReplaceAll(s, "\n", "\n"+indent) // nolint:nlreturn
}

// FormatInfo formats  several lines with a suffix and indent.
func FormatInfo(lines ...string) string {
	return FormatMessage(infoSuffix, indent, lines...)
}
func FormatInfoF(format string, args ...any) string {
	return FormatMessageF(infoSuffix, indent, format, args...)
}

type proxyWriterInfo struct {
	log ILogContext
}

type proxyWriterErr struct {
	log ILogContext
}

func (p *proxyWriterInfo) Write(s []byte) (int, error) {
	return p.log.GetInfoWriter().Write(s)
}

func (p *proxyWriterErr) Write(s []byte) (int, error) {
	return p.log.GetErrorWriter().Write(s)
}

// ToInfoWriter wraps the log context info into a `io.Writer`.
func ToInfoWriter(log ILogContext) io.Writer {
	return &proxyWriterInfo{log: log}
}

// ToErrorWriter wraps the log context error into a `io.Writer`.
func ToErrorWriter(log ILogContext) io.Writer {
	return &proxyWriterErr{log: log}
}
