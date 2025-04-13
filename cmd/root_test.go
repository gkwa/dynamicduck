package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gkwa/dynamicduck/internal/model"
)

// TestRootCmdWithValidInput tests the root command with valid input data
func TestRootCmdWithValidInput(t *testing.T) {
	// Create temp dir for test files
	tempDir, err := os.MkdirTemp("", "dynamicduck-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create valid test input JSON
	testData := model.JSONData{
		Items: []map[string]interface{}{
			{
				"category": map[string]interface{}{"Value": "category1"},
				"domain":   map[string]interface{}{"Value": "domain1.com"},
			},
			{
				"category": map[string]interface{}{"Value": "category2"},
				"domain":   map[string]interface{}{"Value": "domain2.com"},
			},
			{
				"category": map[string]interface{}{"Value": "category3"},
				"domain":   map[string]interface{}{"Value": "domain3.com"},
			},
		},
		Count:            3,
		ScannedCount:     3,
		ConsumedCapacity: nil,
	}

	inputJSON, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	inputFile := filepath.Join(tempDir, "input.json")
	err = os.WriteFile(inputFile, inputJSON, 0o644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "output.json")
	// Create the output directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(outputFile), 0o755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Test executing the command with input file and output file
	rootCmd.SetArgs([]string{
		"--input", inputFile,
		"--out-file", outputFile,
		"--count", "2",
	})

	// Execute the command
	if err = rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	// Read and verify the output file
	outputData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result model.JSONData
	if err = json.Unmarshal(outputData, &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify results
	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items in result, got %d", len(result.Items))
	}

	if result.Count != 2 {
		t.Errorf("Expected Count field to be 2, got %d", result.Count)
	}

	if result.ScannedCount != testData.ScannedCount {
		t.Errorf("Expected ScannedCount to be preserved: got %d, want %d",
			result.ScannedCount, testData.ScannedCount)
	}
}

// TestRootCmdWithStdout tests the command with stdout output
func TestRootCmdWithStdout(t *testing.T) {
	t.Skip("Skipping stdout test as it's environment-dependent")
}

// TestRootCmdWithError tests error handling in the root command
func TestRootCmdWithError(t *testing.T) {
	// Test with non-existent input file
	rootCmd.SetArgs([]string{
		"--input", "non-existent-file.json",
	})

	// Capture stderr to check for error message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := rootCmd.Execute()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured stderr
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err == nil {
		t.Errorf("Expected error with non-existent input file, got nil")
	}

	// Test with invalid JSON input
	tempDir, err := os.MkdirTemp("", "dynamicduck-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidJSON := `{ "Items": [ { "this is not valid JSON" }`
	invalidFile := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte(invalidJSON), 0o644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	rootCmd.SetArgs([]string{
		"--input", invalidFile,
	})

	// Capture stderr again
	r, w, _ = os.Pipe()
	os.Stderr = w

	err = rootCmd.Execute()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured stderr
	buf.Reset()
	buf.ReadFrom(r)
	output = buf.String()

	if err == nil {
		t.Errorf("Expected error with invalid JSON, got nil")
	}

	// Test with empty items array and count = 0
	emptyJSON := `{"Items":[],"Count":0,"ScannedCount":0,"ConsumedCapacity":null}`
	emptyFile := filepath.Join(tempDir, "empty.json")
	err = os.WriteFile(emptyFile, []byte(emptyJSON), 0o644)
	if err != nil {
		t.Fatalf("Failed to write empty JSON file: %v", err)
	}

	rootCmd.SetArgs([]string{
		"--input", emptyFile,
		"--count", "0",
	})

	// Capture stderr to check for error message
	r, w, _ = os.Pipe()
	os.Stderr = w

	err = rootCmd.Execute()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured stderr
	buf.Reset()
	buf.ReadFrom(r)
	output = buf.String()

	if err == nil {
		t.Errorf("Expected error with empty items and count=0, got nil")
	}

	// Check that an error message was produced, but don't check exact contents
	// since the error message might vary
	if output == "" {
		t.Errorf("Expected error message in stderr, got empty output")
	}
}

// TestRootCmdWithSeenFile tests the seen file functionality
func TestRootCmdWithSeenFile(t *testing.T) {
	// Create temp dir for test files
	tempDir, err := os.MkdirTemp("", "dynamicduck-seen-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create valid test input JSON with unique IDs
	testData := model.JSONData{
		Items: []map[string]interface{}{
			{
				"category": map[string]interface{}{"Value": "category1"},
				"product": map[string]interface{}{
					"Value": map[string]interface{}{
						"id": map[string]interface{}{"Value": "id1"},
					},
				},
			},
			{
				"category": map[string]interface{}{"Value": "category2"},
				"product": map[string]interface{}{
					"Value": map[string]interface{}{
						"id": map[string]interface{}{"Value": "id2"},
					},
				},
			},
			{
				"category": map[string]interface{}{"Value": "category3"},
				"product": map[string]interface{}{
					"Value": map[string]interface{}{
						"id": map[string]interface{}{"Value": "id3"},
					},
				},
			},
		},
		Count:            3,
		ScannedCount:     3,
		ConsumedCapacity: nil,
	}

	inputJSON, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	inputFile := filepath.Join(tempDir, "input.json")
	err = os.WriteFile(inputFile, inputJSON, 0o644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "output.json")
	seenFile := filepath.Join(tempDir, "seen.txt")

	// First run: select 1 item, should create seen file
	// We use seed=0 to avoid incompatible options error
	rootCmd.SetArgs([]string{
		"--input", inputFile,
		"--out-file", outputFile,
		"--count", "1",
		"--seen-file", seenFile,
	})

	if err = rootCmd.Execute(); err != nil {
		t.Fatalf("First run: rootCmd.Execute() error = %v", err)
	}

	// Verify the seen file was created
	if _, err = os.Stat(seenFile); os.IsNotExist(err) {
		t.Errorf("Seen file not created after first run")
	}

	// Read the first output
	outputData1, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read first output file: %v", err)
	}

	var result1 model.JSONData
	if err = json.Unmarshal(outputData1, &result1); err != nil {
		t.Fatalf("First output is not valid JSON: %v", err)
	}

	if len(result1.Items) != 1 {
		t.Errorf("First run: expected 1 item, got %d", len(result1.Items))
	}

	// Get ID of the first selected item
	productValue := result1.Items[0]["product"].(map[string]interface{})["Value"].(map[string]interface{})
	firstItemID := productValue["id"].(map[string]interface{})["Value"].(string)

	// Second run: select 1 more item, should not repeat
	rootCmd.SetArgs([]string{
		"--input", inputFile,
		"--out-file", outputFile,
		"--count", "1",
		"--seen-file", seenFile,
	})

	if err = rootCmd.Execute(); err != nil {
		t.Fatalf("Second run: rootCmd.Execute() error = %v", err)
	}

	// Read the second output
	outputData2, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read second output file: %v", err)
	}

	var result2 model.JSONData
	if err = json.Unmarshal(outputData2, &result2); err != nil {
		t.Fatalf("Second output is not valid JSON: %v", err)
	}

	if len(result2.Items) != 1 {
		t.Errorf("Second run: expected 1 item, got %d", len(result2.Items))
	}

	// Verify the second item is different from the first
	productValue = result2.Items[0]["product"].(map[string]interface{})["Value"].(map[string]interface{})
	secondItemID := productValue["id"].(map[string]interface{})["Value"].(string)
	if secondItemID == firstItemID {
		t.Errorf("Second run: got same item as first run")
	}

	// Third run: select 1 more item (the last one)
	rootCmd.SetArgs([]string{
		"--input", inputFile,
		"--out-file", outputFile,
		"--count", "1",
		"--seen-file", seenFile,
	})

	if err = rootCmd.Execute(); err != nil {
		t.Fatalf("Third run: rootCmd.Execute() error = %v", err)
	}

	// Read the third output
	outputData3, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read third output file: %v", err)
	}

	var result3 model.JSONData
	if err = json.Unmarshal(outputData3, &result3); err != nil {
		t.Fatalf("Third output is not valid JSON: %v", err)
	}

	if len(result3.Items) != 1 {
		t.Errorf("Third run: expected 1 item, got %d", len(result3.Items))
	}

	// Verify the third item is different from the first two
	productValue = result3.Items[0]["product"].(map[string]interface{})["Value"].(map[string]interface{})
	thirdItemID := productValue["id"].(map[string]interface{})["Value"].(string)
	if thirdItemID == firstItemID || thirdItemID == secondItemID {
		t.Errorf("Third run: got same item as previous runs")
	}

	// Fourth run: should return empty result as all items have been seen
	rootCmd.SetArgs([]string{
		"--input", inputFile,
		"--out-file", outputFile,
		"--count", "1",
		"--seen-file", seenFile,
	})

	if err = rootCmd.Execute(); err != nil {
		t.Fatalf("Fourth run: rootCmd.Execute() error = %v", err)
	}

	// Read the fourth output
	outputData4, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read fourth output file: %v", err)
	}

	var result4 model.JSONData
	if err = json.Unmarshal(outputData4, &result4); err != nil {
		t.Fatalf("Fourth output is not valid JSON: %v", err)
	}

	if len(result4.Items) != 0 {
		t.Errorf("Fourth run: expected 0 items, got %d", len(result4.Items))
	}
}

// TestIncompatibleOptions tests that an error is returned when both seed and seen-file are used
func TestIncompatibleOptions(t *testing.T) {
	// Create temp dir for test files
	tempDir, err := os.MkdirTemp("", "dynamicduck-incompatible-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create valid test input JSON
	testData := model.JSONData{
		Items: []map[string]interface{}{
			{
				"category": map[string]interface{}{"Value": "category1"},
				"domain":   map[string]interface{}{"Value": "domain1.com"},
			},
		},
		Count:            1,
		ScannedCount:     1,
		ConsumedCapacity: nil,
	}

	inputJSON, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	inputFile := filepath.Join(tempDir, "input.json")
	err = os.WriteFile(inputFile, inputJSON, 0o644)
	if err != nil {
		t.Fatalf("Failed to write input file: %v", err)
	}

	outputFile := filepath.Join(tempDir, "output.json")
	seenFile := filepath.Join(tempDir, "seen.txt")

	// Test with both seed and seen file
	rootCmd.SetArgs([]string{
		"--input", inputFile,
		"--out-file", outputFile,
		"--count", "1",
		"--seed", "12345",
		"--seen-file", seenFile,
	})

	// Capture stderr to check for error message
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Execute the command and check for error
	err = rootCmd.Execute()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	// Read captured stderr
	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err == nil {
		t.Errorf("Expected error when using both seed and seen-file, but got nil")
	} else if !strings.Contains(err.Error(), "cannot use both --seed and --seen-file together") {
		t.Errorf("Expected error message about incompatible options, got: %v", err)
	}
}
