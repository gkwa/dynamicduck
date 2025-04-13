package sampler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gkwa/dynamicduck/internal/model"
)

func TestSampleItems(t *testing.T) {
	// Create test data
	testData := &model.JSONData{
		Items: []map[string]interface{}{
			{
				"category": map[string]interface{}{"Value": "item1"},
			},
			{
				"category": map[string]interface{}{"Value": "item2"},
			},
			{
				"category": map[string]interface{}{"Value": "item3"},
			},
			{
				"category": map[string]interface{}{"Value": "item4"},
			},
			{
				"category": map[string]interface{}{"Value": "item5"},
			},
		},
		Count:            5,
		ScannedCount:     5,
		ConsumedCapacity: nil,
	}

	tests := []struct {
		name        string
		data        *model.JSONData
		count       int
		seed        int64
		seenFile    string
		wantItemLen int
		wantCount   int
		wantErr     bool
	}{
		{
			name:        "Sample 3 from 5",
			data:        testData,
			count:       3,
			seed:        0, // Use time-based seed
			seenFile:    "",
			wantItemLen: 3,
			wantCount:   3,
			wantErr:     false,
		},
		{
			name:        "Sample more than available",
			data:        testData,
			count:       10,
			seed:        0,
			seenFile:    "",
			wantItemLen: 5, // Should return all 5 items
			wantCount:   5, // Count should be adjusted to actual item count
			wantErr:     false,
		},
		{
			name: "Sample from empty array",
			data: &model.JSONData{
				Items:            []map[string]interface{}{},
				Count:            0,
				ScannedCount:     0,
				ConsumedCapacity: nil,
			},
			count:       5,
			seed:        0,
			seenFile:    "",
			wantItemLen: 0,
			wantCount:   0, // Count should be 0 for empty result
			wantErr:     false,
		},
		{
			name:        "Sample 0 items",
			data:        testData,
			count:       0,
			seed:        0,
			seenFile:    "",
			wantItemLen: 0,
			wantCount:   0,
			wantErr:     false,
		},
		{
			name:        "Incompatible options: seed and seen file",
			data:        testData,
			count:       3,
			seed:        123,
			seenFile:    "some-file.txt", // Both seed and seenFile are set
			wantItemLen: 0,
			wantCount:   0,
			wantErr:     true, // Should produce an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SampleItems(tt.data, tt.count, tt.seed, tt.seenFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("SampleItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expect an error, don't check the results further
			if tt.wantErr {
				return
			}

			// Check result length
			if len(got.Items) != tt.wantItemLen {
				t.Errorf("SampleItems() len = %v, want %v", len(got.Items), tt.wantItemLen)
			}

			// Check metadata is preserved
			if got.ScannedCount != tt.data.ScannedCount {
				t.Errorf("SampleItems() didn't preserve ScannedCount = %v, want %v", got.ScannedCount, tt.data.ScannedCount)
			}

			// Check Count is updated correctly
			if got.Count != tt.wantCount {
				t.Errorf("SampleItems() Count = %v, want %v", got.Count, tt.wantCount)
			}
		})
	}
}

// TestIncompatibleOptions specifically tests the seed + seen file error case
func TestIncompatibleOptions(t *testing.T) {
	testData := &model.JSONData{
		Items: []map[string]interface{}{
			{
				"category": map[string]interface{}{"Value": "item1"},
			},
		},
		Count:            1,
		ScannedCount:     1,
		ConsumedCapacity: nil,
	}

	// Create a temporary file for the test
	tempDir, err := os.MkdirTemp("", "sampler-incompatible-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	seenFile := filepath.Join(tempDir, "seen.txt")

	// Test with both seed and seen file set
	_, err = SampleItems(testData, 1, 123, seenFile)
	if err == nil {
		t.Errorf("Expected error when using both seed and seen file, but got nil")
	} else if !strings.Contains(err.Error(), "cannot use both --seed and --seen-file together") {
		t.Errorf("Expected error message about incompatible options, got: %v", err)
	}
}

// TestDeterministicSampling checks that using the same seed produces the same results
func TestDeterministicSampling(t *testing.T) {
	// Create test data with 100 items
	items := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]interface{}{
			"category": map[string]interface{}{"Value": i},
		}
	}

	testData := &model.JSONData{
		Items:            items,
		Count:            100,
		ScannedCount:     100,
		ConsumedCapacity: nil,
	}

	// Sample using a fixed seed
	const seed = 12345
	const sampleSize = 10

	// Run sampling twice with the same seed
	result1, err := SampleItems(testData, sampleSize, seed, "")
	if err != nil {
		t.Fatalf("SampleItems() failed: %v", err)
	}

	result2, err := SampleItems(testData, sampleSize, seed, "")
	if err != nil {
		t.Fatalf("SampleItems() failed: %v", err)
	}

	// Check that both results have the same items in the same order
	if len(result1.Items) != len(result2.Items) {
		t.Fatalf("Results have different lengths: %d vs %d", len(result1.Items), len(result2.Items))
	}

	for i := 0; i < len(result1.Items); i++ {
		cat1 := result1.Items[i]["category"].(map[string]interface{})["Value"]
		cat2 := result2.Items[i]["category"].(map[string]interface{})["Value"]
		if cat1 != cat2 {
			t.Errorf("Item %d differs between samples: %v vs %v", i, cat1, cat2)
		}
	}

	// Run sampling with a different seed
	const differentSeed = 54321
	result3, err := SampleItems(testData, sampleSize, differentSeed, "")
	if err != nil {
		t.Fatalf("SampleItems() failed: %v", err)
	}

	// Check that the results are different (this could theoretically fail by chance,
	// but is extremely unlikely with 100 items and 10 samples)
	allSame := true
	for i := 0; i < len(result1.Items); i++ {
		cat1 := result1.Items[i]["category"].(map[string]interface{})["Value"]
		cat3 := result3.Items[i]["category"].(map[string]interface{})["Value"]
		if cat1 != cat3 {
			allSame = false
			break
		}
	}

	if allSame {
		t.Errorf("SampleItems() with different seeds produced identical results")
	}
}

// TestRandomness checks that the sampling is actually random when no seed is provided
func TestRandomness(t *testing.T) {
	// Create test data with 100 items
	items := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]interface{}{
			"category": map[string]interface{}{"Value": i},
		}
	}

	testData := &model.JSONData{
		Items:            items,
		Count:            100,
		ScannedCount:     100,
		ConsumedCapacity: nil,
	}

	// Sample 10 items multiple times with time-based seed (0)
	const sampleSize = 10
	const iterations = 5

	// Track results to check for randomness
	allSamples := make([][]interface{}, iterations)

	for i := 0; i < iterations; i++ {
		result, err := SampleItems(testData, sampleSize, 0, "") // Use time-based seed
		if err != nil {
			t.Fatalf("SampleItems() failed: %v", err)
		}

		// Extract category values
		sample := make([]interface{}, sampleSize)
		for j := 0; j < sampleSize; j++ {
			sample[j] = result.Items[j]["category"].(map[string]interface{})["Value"]
		}
		allSamples[i] = sample
	}

	// Check that at least some samples differ
	// This is not a perfect test for randomness but should catch obvious problems
	allSame := true
	firstSample := allSamples[0]
	for i := 1; i < iterations; i++ {
		different := false
		for j := 0; j < sampleSize; j++ {
			if firstSample[j] != allSamples[i][j] {
				different = true
				break
			}
		}
		if different {
			allSame = false
			break
		}
	}

	if allSame {
		t.Errorf("All samples were identical, suggesting non-random sampling")
	}
}

// TestSampleItemsWithSeenFile tests the seen file functionality
func TestSampleItemsWithSeenFile(t *testing.T) {
	// Create test data
	testData := &model.JSONData{
		Items: []map[string]interface{}{
			{
				"category": map[string]interface{}{"Value": "item1"},
				"product": map[string]interface{}{
					"Value": map[string]interface{}{
						"id": map[string]interface{}{"Value": "id1"},
					},
				},
			},
			{
				"category": map[string]interface{}{"Value": "item2"},
				"product": map[string]interface{}{
					"Value": map[string]interface{}{
						"id": map[string]interface{}{"Value": "id2"},
					},
				},
			},
			{
				"category": map[string]interface{}{"Value": "item3"},
				"product": map[string]interface{}{
					"Value": map[string]interface{}{
						"id": map[string]interface{}{"Value": "id3"},
					},
				},
			},
			{
				"category": map[string]interface{}{"Value": "item4"},
				"product": map[string]interface{}{
					"Value": map[string]interface{}{
						"id": map[string]interface{}{"Value": "id4"},
					},
				},
			},
			{
				"category": map[string]interface{}{"Value": "item5"},
				"product": map[string]interface{}{
					"Value": map[string]interface{}{
						"id": map[string]interface{}{"Value": "id5"},
					},
				},
			},
		},
		Count:            5,
		ScannedCount:     5,
		ConsumedCapacity: nil,
	}

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "sampler-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a seen file path
	seenFile := filepath.Join(tempDir, "seen.txt")

	// First sampling: should get 2 items with no seen items yet
	// Note: using 0 for seed to avoid incompatible options error
	result1, err := SampleItems(testData, 2, 0, seenFile)
	if err != nil {
		t.Fatalf("First SampleItems() failed: %v", err)
	}

	if len(result1.Items) != 2 {
		t.Errorf("First sample: expected 2 items, got %d", len(result1.Items))
	}

	// Verify the seen file was created
	if _, err := os.Stat(seenFile); os.IsNotExist(err) {
		t.Errorf("Seen file not created after first sample")
	}

	// Extract IDs from first sample for verification
	firstSampleIDs := make([]string, len(result1.Items))
	for i, item := range result1.Items {
		productValue := item["product"].(map[string]interface{})["Value"].(map[string]interface{})
		idValue := productValue["id"].(map[string]interface{})["Value"].(string)
		firstSampleIDs[i] = idValue
	}

	// Second sampling: should get 2 more items that are different from the first
	result2, err := SampleItems(testData, 2, 0, seenFile)
	if err != nil {
		t.Fatalf("Second SampleItems() failed: %v", err)
	}

	if len(result2.Items) != 2 {
		t.Errorf("Second sample: expected 2 items, got %d", len(result2.Items))
	}

	// Verify the second sample has different items
	for _, item := range result2.Items {
		productValue := item["product"].(map[string]interface{})["Value"].(map[string]interface{})
		idValue := productValue["id"].(map[string]interface{})["Value"].(string)

		for _, firstID := range firstSampleIDs {
			if idValue == firstID {
				t.Errorf("Second sample included already seen item with ID: %s", idValue)
			}
		}
	}

	// Third sampling: should get the last item
	result3, err := SampleItems(testData, 2, 0, seenFile)
	if err != nil {
		t.Fatalf("Third SampleItems() failed: %v", err)
	}

	if len(result3.Items) != 1 {
		t.Errorf("Third sample: expected 1 item, got %d", len(result3.Items))
	}

	// Fourth sampling: should fail as all items have been seen
	result4, err := SampleItems(testData, 2, 0, seenFile)
	if err == nil {
		t.Errorf("Fourth sample: expected error but got none")
	}

	if len(result4.Items) != 0 {
		t.Errorf("Fourth sample: expected 0 items, got %d", len(result4.Items))
	}
}
