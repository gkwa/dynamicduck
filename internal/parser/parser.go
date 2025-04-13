package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gkwa/dynamicduck/internal/model"
)

// ReadInput reads data from stdin or a file
func ReadInput(inputFile string) ([]byte, error) {
	var reader io.Reader

	if inputFile == "" {
		// Read from stdin
		reader = os.Stdin
	} else {
		// Read from file
		file, err := os.Open(inputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()
		reader = file
	}

	return io.ReadAll(reader)
}

// ParseJSON parses the JSON data into the model structure
func ParseJSON(data []byte) (*model.JSONData, error) {
	var jsonData model.JSONData
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}
	return &jsonData, nil
}

// GenerateOutput generates the output JSON
func GenerateOutput(data *model.JSONData) ([]byte, error) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return output, nil
}
