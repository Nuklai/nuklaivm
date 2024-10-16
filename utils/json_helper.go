// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"encoding/json"
)

// Function to convert a JSON string to map[string]string
func BytesToMap(jsonBytes []byte) (map[string]string, error) {
	// Create a map to hold the unmarshaled JSON
	var result map[string]string

	// Unmarshal the JSON string into the map
	err := json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, err
	}

	// Return the map and any error encountered
	return result, nil
}

// Function to convert a map[string]string to a JSON string
func MapToBytes(inputMap map[string]string) ([]byte, error) {
	// Marshal the map into a JSON string
	jsonBytes, err := json.Marshal(inputMap)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}
