package logger

import (
	"fmt"
	"os"
)

// Logger provides logging functionality with different verbosity levels
type Logger struct {
	verbosityLevel int
}

// NewLogger creates a new logger with the specified verbosity level
func NewLogger(verbosityLevel int) *Logger {
	return &Logger{
		verbosityLevel: verbosityLevel,
	}
}

// Error logs error messages (always shown)
func (l *Logger) Error(message string) {
	fmt.Fprintln(os.Stderr, "ERROR:", message)
}

// Verbose logs basic operation messages (shown with -v or higher)
func (l *Logger) Verbose(message string) {
	if l.verbosityLevel >= 1 {
		fmt.Fprintln(os.Stderr, "INFO:", message)
	}
}

// VeryVerbose logs detailed operation messages (shown with -vv or higher)
func (l *Logger) VeryVerbose(message string) {
	if l.verbosityLevel >= 2 {
		fmt.Fprintln(os.Stderr, "DETAIL:", message)
	}
}

// Debug logs debug information (shown with -vvv)
func (l *Logger) Debug(message string) {
	if l.verbosityLevel >= 3 {
		fmt.Fprintln(os.Stderr, "DEBUG:", message)
	}
}
