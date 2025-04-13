package model

// JSONData represents the overall JSON structure
type JSONData struct {
	Items            []map[string]interface{} `json:"Items"`
	Count            int                      `json:"Count"`
	ScannedCount     int                      `json:"ScannedCount"`
	ConsumedCapacity interface{}              `json:"ConsumedCapacity"`
}
