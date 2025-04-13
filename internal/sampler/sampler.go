package sampler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gkwa/dynamicduck/internal/model"
)

// SampleItems randomly samples n items from the input data
// If seed is 0, uses the current time as the seed
// If seed is non-zero, uses the provided seed for deterministic results
// If seenFile is not empty, will read seen IDs and avoid selecting them
func SampleItems(data *model.JSONData, n int, seed int64, seenFile string) (*model.JSONData, error) {
	// Check for incompatible options
	if seed != 0 && seenFile != "" {
		return nil, fmt.Errorf("cannot use both --seed and --seen-file together, as they serve different purposes")
	}

	// Create a copy of the input data
	result := &model.JSONData{
		ScannedCount:     data.ScannedCount,
		ConsumedCapacity: data.ConsumedCapacity,
	}

	// Handle edge cases
	totalItems := len(data.Items)
	if totalItems == 0 {
		// If there are no items, set Count to 0
		result.Count = 0
		result.Items = []map[string]interface{}{}
		return result, nil
	}

	// Load seen items if a seen file is provided
	seenItems := make(map[string]bool)
	var seenCount int
	if seenFile != "" {
		var err error
		seenItems, err = loadSeenItems(seenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load seen items: %w", err)
		}
		seenCount = len(seenItems)
	}

	// Create a list of item indices that haven't been seen
	unseenIndices := make([]int, 0, totalItems)
	unseenItems := make([]map[string]interface{}, 0, totalItems)

	for i, item := range data.Items {
		// Generate an ID for the item
		itemID := generateItemID(item)

		// Skip if this item has been seen before
		if seenItems[itemID] {
			continue
		}

		unseenIndices = append(unseenIndices, i)
		unseenItems = append(unseenItems, item)
	}

	// If all items have been seen, we can:
	// 1. Return an empty result
	// 2. Reset and use all items
	// 3. Allow duplicates with a warning
	// For now, we'll use option 1
	if len(unseenItems) == 0 {
		result.Count = 0
		result.Items = []map[string]interface{}{}
		return result, fmt.Errorf("all items have been seen already (seen count: %d)", seenCount)
	}

	// If n is greater than or equal to the total number of unseen items, use all unseen items
	availableCount := len(unseenItems)
	if n >= availableCount {
		result.Items = make([]map[string]interface{}, availableCount)
		copy(result.Items, unseenItems)
		result.Count = availableCount

		// Update seen items file with the newly sampled items
		if seenFile != "" {
			for _, item := range unseenItems {
				seenItems[generateItemID(item)] = true
			}
			if err := saveSeenItems(seenFile, seenItems); err != nil {
				return nil, fmt.Errorf("failed to update seen items file: %w", err)
			}
		}

		return result, nil
	}

	// Initialize random number generator
	var r *rand.Rand
	if seed == 0 {
		// Use current time as seed if not provided
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		// Use provided seed for deterministic results
		r = rand.New(rand.NewSource(seed))
	}

	// Randomly select n items using Fisher-Yates shuffle algorithm on the unseen items
	result.Items = make([]map[string]interface{}, n)
	shuffledIndices := make([]int, availableCount)

	// Create copy of the indices for shuffling
	for i := 0; i < availableCount; i++ {
		shuffledIndices[i] = i
	}

	// Shuffle the first n items
	for i := 0; i < n; i++ {
		// Generate a random index between i and availableCount-1
		j := r.Intn(availableCount-i) + i

		// Swap the indices at positions i and j
		shuffledIndices[i], shuffledIndices[j] = shuffledIndices[j], shuffledIndices[i]

		// Add the item at shuffled position i to the result
		result.Items[i] = unseenItems[shuffledIndices[i]]

		// Mark this item as seen
		if seenFile != "" {
			seenItems[generateItemID(unseenItems[shuffledIndices[i]])] = true
		}
	}

	// Update the seen items file
	if seenFile != "" {
		if err := saveSeenItems(seenFile, seenItems); err != nil {
			return nil, fmt.Errorf("failed to update seen items file: %w", err)
		}
	}

	// Set the Count field to the number of sampled items
	result.Count = n

	return result, nil
}

// generateItemID creates a unique ID for an item based on its properties
// This is used to track which items have been seen before
func generateItemID(item map[string]interface{}) string {
	// Try to find an ID in the item or its nested structures

	// First, look for a direct id or ID field
	if id, ok := extractStringValue(item, "id"); ok && id != "" {
		return id
	}
	if id, ok := extractStringValue(item, "ID"); ok && id != "" {
		return id
	}

	// Look for product.id or similar nested paths
	if product, ok := item["product"].(map[string]interface{}); ok {
		if productValue, ok := product["Value"].(map[string]interface{}); ok {
			if id, ok := extractStringValue(productValue, "id"); ok && id != "" {
				return id
			}
		}
	}

	// Check for timestamp
	if timestamp, ok := extractStringValue(item, "timestamp"); ok && timestamp != "" {
		return timestamp
	}

	// Create an ID from other properties
	var parts []string

	// Try to extract common fields
	if category, ok := extractStringValue(item, "category"); ok && category != "" {
		parts = append(parts, "cat:"+category)
	}

	if domain, ok := extractStringValue(item, "domain"); ok && domain != "" {
		parts = append(parts, "dom:"+domain)
	}

	// If product name is available, try to use it
	if product, ok := item["product"].(map[string]interface{}); ok {
		if productValue, ok := product["Value"].(map[string]interface{}); ok {
			if name, ok := extractStringValue(productValue, "name"); ok && name != "" {
				parts = append(parts, "name:"+name)
			}
		}
	}

	// If we have no parts, create a hash of the JSON
	if len(parts) == 0 {
		jsonBytes, err := json.Marshal(item)
		if err == nil {
			return string(jsonBytes)
		}
		return fmt.Sprintf("%v", item)
	}

	return strings.Join(parts, "|")
}

// extractStringValue tries to extract a string value from a potentially nested structure
func extractStringValue(m map[string]interface{}, key string) (string, bool) {
	// Direct string value
	if val, ok := m[key].(string); ok {
		return val, true
	}

	// Value wrapper: {"Value": "some-value"}
	if wrapper, ok := m[key].(map[string]interface{}); ok {
		if val, ok := wrapper["Value"].(string); ok {
			return val, true
		}
	}

	return "", false
}

// loadSeenItems loads the list of seen items from a file
// Each line in the file is treated as a unique item ID
func loadSeenItems(filePath string) (map[string]bool, error) {
	seenItems := make(map[string]bool)

	// If the file doesn't exist, return an empty map
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return seenItems, nil
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open seen items file: %w", err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		id := strings.TrimSpace(scanner.Text())
		if id != "" {
			seenItems[id] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading seen items file: %w", err)
	}

	return seenItems, nil
}

// saveSeenItems saves the list of seen items to a file
func saveSeenItems(filePath string, seenItems map[string]bool) error {
	// Create the file (or truncate if it exists)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create seen items file: %w", err)
	}
	defer file.Close()

	// Write each seen item ID to the file
	writer := bufio.NewWriter(file)
	for id := range seenItems {
		if _, err := writer.WriteString(id + "\n"); err != nil {
			return fmt.Errorf("failed to write to seen items file: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush seen items file: %w", err)
	}

	return nil
}
