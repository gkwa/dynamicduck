package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	// Save and restore stderr
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()

	tests := []struct {
		name           string
		verbosityLevel int
		message        string
		logFunc        func(l *Logger, msg string)
		shouldContain  bool
	}{
		{
			name:           "Error with verbosity 0",
			verbosityLevel: 0,
			message:        "test error",
			logFunc:        func(l *Logger, msg string) { l.Error(msg) },
			shouldContain:  true, // Errors should always be logged
		},
		{
			name:           "Verbose with verbosity 0",
			verbosityLevel: 0,
			message:        "test verbose",
			logFunc:        func(l *Logger, msg string) { l.Verbose(msg) },
			shouldContain:  false, // Verbose shouldn't log at level 0
		},
		{
			name:           "Verbose with verbosity 1",
			verbosityLevel: 1,
			message:        "test verbose",
			logFunc:        func(l *Logger, msg string) { l.Verbose(msg) },
			shouldContain:  true, // Verbose should log at level 1
		},
		{
			name:           "VeryVerbose with verbosity 1",
			verbosityLevel: 1,
			message:        "test very verbose",
			logFunc:        func(l *Logger, msg string) { l.VeryVerbose(msg) },
			shouldContain:  false, // VeryVerbose shouldn't log at level 1
		},
		{
			name:           "VeryVerbose with verbosity 2",
			verbosityLevel: 2,
			message:        "test very verbose",
			logFunc:        func(l *Logger, msg string) { l.VeryVerbose(msg) },
			shouldContain:  true, // VeryVerbose should log at level 2
		},
		{
			name:           "Debug with verbosity 2",
			verbosityLevel: 2,
			message:        "test debug",
			logFunc:        func(l *Logger, msg string) { l.Debug(msg) },
			shouldContain:  false, // Debug shouldn't log at level 2
		},
		{
			name:           "Debug with verbosity 3",
			verbosityLevel: 3,
			message:        "test debug",
			logFunc:        func(l *Logger, msg string) { l.Debug(msg) },
			shouldContain:  true, // Debug should log at level 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Create logger and call the log function
			logger := NewLogger(tt.verbosityLevel)
			tt.logFunc(logger, tt.message)

			// Close the writer to get all output
			w.Close()

			// Read the captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check if the output contains the message
			if tt.shouldContain && !strings.Contains(output, tt.message) {
				t.Errorf("Log output should contain message '%s', but got: %s", tt.message, output)
			}
			if !tt.shouldContain && strings.Contains(output, tt.message) {
				t.Errorf("Log output should not contain message '%s', but got: %s", tt.message, output)
			}
		})
	}
}
