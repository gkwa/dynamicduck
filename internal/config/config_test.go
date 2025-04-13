package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	// Check default values
	if config.Count != 10 {
		t.Errorf("NewConfig() Count = %v, want %v", config.Count, 10)
	}

	if config.InputFile != "" {
		t.Errorf("NewConfig() InputFile = %v, want %v", config.InputFile, "")
	}

	if config.OutputFile != "" {
		t.Errorf("NewConfig() OutputFile = %v, want %v", config.OutputFile, "")
	}

	if config.VerbosityLevel != 0 {
		t.Errorf("NewConfig() VerbosityLevel = %v, want %v", config.VerbosityLevel, 0)
	}
}
