package cmd

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gkwa/dynamicduck/internal/model"
)

// Integration test using the actual binary
// This test requires the binary to be built first.
// These tests will be skipped if the binary isn't found.
func TestIntegration(t *testing.T) {
	// Check if the binary exists
	binaryPath, err := findBinary()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// Create temp dir for test files
	tempDir, err := os.MkdirTemp("", "dynamicduck-integration")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test data with 100 items
	testData := createTestData(100)
	inputFile := filepath.Join(tempDir, "input.json")
	outputFile := filepath.Join(tempDir, "output.json")

	// Write test data to file
	inputJSON, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}
	err = os.WriteFile(inputFile, inputJSON, 0o644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	// Test cases
	tests := []struct {
		name        string
		args        []string
		wantCount   int
		wantSuccess bool
	}{
		{
			name:        "Basic usage",
			args:        []string{"--input", inputFile, "--out-file", outputFile, "--count", "20"},
			wantCount:   20,
			wantSuccess: true,
		},
		{
			name:        "Verbose output",
			args:        []string{"--input", inputFile, "--out-file", outputFile, "--count", "5", "-v"},
			wantCount:   5,
			wantSuccess: true,
		},
		{
			name:        "Count too large",
			args:        []string{"--input", inputFile, "--out-file", outputFile, "--count", "200"},
			wantCount:   100, // Should adjust to actual count
			wantSuccess: true,
		},
		{
			name:        "Invalid count",
			args:        []string{"--input", inputFile, "--out-file", outputFile, "--count", "0"},
			wantCount:   0,
			wantSuccess: false,
		},
		{
			name:        "Non-existent input file",
			args:        []string{"--input", "non-existent.json", "--out-file", outputFile},
			wantCount:   0,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run the command
			cmd := exec.Command(binaryPath, tt.args...)
			stderr, err := cmd.StderrPipe()
			if err != nil {
				t.Fatalf("Failed to get stderr pipe: %v", err)
			}

			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start command: %v", err)
			}

			// Read stderr for debugging
			stderrBytes, _ := io.ReadAll(stderr)
			stderrOutput := string(stderrBytes)

			err = cmd.Wait()
			if tt.wantSuccess && err != nil {
				t.Errorf("Command failed but expected success: %v\nStderr: %s", err, stderrOutput)
				return
			}
			if !tt.wantSuccess && err == nil {
				t.Errorf("Command succeeded but expected failure")
				return
			}

			// If we expect success, verify the output
			if tt.wantSuccess {
				outputData, err := os.ReadFile(outputFile)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				var result model.JSONData
				if err := json.Unmarshal(outputData, &result); err != nil {
					t.Fatalf("Output is not valid JSON: %v", err)
				}

				// Verify result
				if len(result.Items) != tt.wantCount {
					t.Errorf("Expected %d items in result, got %d", tt.wantCount, len(result.Items))
				}

				if result.Count != tt.wantCount {
					t.Errorf("Expected Count field to be %d, got %d", tt.wantCount, result.Count)
				}

				// Check that metadata is preserved
				if result.ScannedCount != testData.ScannedCount {
					t.Errorf("ScannedCount not preserved: got %d, want %d",
						result.ScannedCount, testData.ScannedCount)
				}
			}
		})
	}
}

// Helper function to find the binary
func findBinary() (string, error) {
	// Try common locations
	locations := []string{
		"../dynamicduck",        // Binary in parent directory
		"../bin/dynamicduck",    // Binary in bin directory
		"../../bin/dynamicduck", // Binary in parent's bin directory
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", os.ErrNotExist
}

// Helper function to create test data
func createTestData(count int) *model.JSONData {
	items := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		items[i] = map[string]interface{}{
			"category": map[string]interface{}{"Value": "category" + string(rune(i+1))},
			"domain":   map[string]interface{}{"Value": "domain" + string(rune(i+1)) + ".com"},
		}
	}

	return &model.JSONData{
		Items:            items,
		Count:            count,
		ScannedCount:     count,
		ConsumedCapacity: nil,
	}
}
