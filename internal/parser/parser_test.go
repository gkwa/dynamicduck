package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gkwa/dynamicduck/internal/model"
)

func TestParseJSON(t *testing.T) {
	// Valid JSON
	validJSON := `{
		"Items": [
			{
				"category": { "Value": "test category" },
				"domain": { "Value": "test.com" },
				"product": {
					"Value": {
						"name": { "Value": "Test Product" },
						"price": { "Value": "$9.99" }
					}
				}
			}
		],
		"Count": 1,
		"ScannedCount": 1,
		"ConsumedCapacity": null
	}`

	// Invalid JSON
	invalidJSON := `{ "Items": [ { "incomplete": }`

	// Empty but valid JSON
	emptyJSON := `{ "Items": [], "Count": 0, "ScannedCount": 0, "ConsumedCapacity": null }`

	tests := []struct {
		name    string
		data    []byte
		want    *model.JSONData
		wantErr bool
	}{
		{
			name: "Valid JSON",
			data: []byte(validJSON),
			want: &model.JSONData{
				Items: []map[string]interface{}{
					{
						"category": map[string]interface{}{"Value": "test category"},
						"domain":   map[string]interface{}{"Value": "test.com"},
						"product": map[string]interface{}{
							"Value": map[string]interface{}{
								"name":  map[string]interface{}{"Value": "Test Product"},
								"price": map[string]interface{}{"Value": "$9.99"},
							},
						},
					},
				},
				Count:            1,
				ScannedCount:     1,
				ConsumedCapacity: nil,
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			data:    []byte(invalidJSON),
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty JSON",
			data: []byte(emptyJSON),
			want: &model.JSONData{
				Items:            []map[string]interface{}{},
				Count:            0,
				ScannedCount:     0,
				ConsumedCapacity: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateOutput(t *testing.T) {
	// Create test data
	testData := &model.JSONData{
		Items: []map[string]interface{}{
			{
				"category": map[string]interface{}{"Value": "test category"},
				"domain":   map[string]interface{}{"Value": "test.com"},
			},
		},
		Count:            1,
		ScannedCount:     1,
		ConsumedCapacity: nil,
	}

	// Generate output
	output, err := GenerateOutput(testData)
	if err != nil {
		t.Fatalf("GenerateOutput() error = %v", err)
	}

	// Parse the output back to verify it's valid JSON
	var parsed model.JSONData
	err = json.Unmarshal(output, &parsed)
	if err != nil {
		t.Fatalf("Generated output is not valid JSON: %v", err)
	}

	// Verify the parsed data matches the original
	if !reflect.DeepEqual(&parsed, testData) {
		t.Errorf("Generated output doesn't match input when parsed: got %v, want %v", &parsed, testData)
	}
}

func TestReadInput(t *testing.T) {
	// Test reading from file
	tempDir, err := os.MkdirTemp("", "parser-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testContent := "test content"
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test reading from file
	data, err := ReadInput(testFile)
	if err != nil {
		t.Fatalf("ReadInput() error = %v", err)
	}
	if string(data) != testContent {
		t.Errorf("ReadInput() = %v, want %v", string(data), testContent)
	}

	// Test with non-existent file
	_, err = ReadInput("non-existent-file.txt")
	if err == nil {
		t.Errorf("ReadInput() with non-existent file should have failed")
	}
}
