package config

// Config holds the application configuration
type Config struct {
	Count          int    // Number of items to select
	InputFile      string // Input file path (empty for stdin)
	OutputFile     string // Output file path (empty for stdout)
	VerbosityLevel int    // Verbosity level (0-3)
	Seed           int64  // Seed for random number generation (0 = use time)
	SeenFile       string // File to track seen items to avoid duplicates
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		Count:          10,
		InputFile:      "",
		OutputFile:     "",
		VerbosityLevel: 0,
		Seed:           0,  // Default to 0 (use time-based seed)
		SeenFile:       "", // Default to empty (don't track seen items)
	}
}
