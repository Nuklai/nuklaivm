// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

// utils.go
package integration

// Function to compare two map[string]string
func mapsEqual(map1, map2 map[string]string) bool {
	// First, compare lengths
	if len(map1) != len(map2) {
		return false
	}

	// Now, iterate through map1 and check if all keys and values are the same in map2
	for key, value1 := range map1 {
		// Check if map2 has the key
		value2, ok := map2[key]
		if !ok {
			// If map2 does not have the key, the maps are not equal
			return false
		}

		// Check if values are the same
		if value1 != value2 {
			return false
		}
	}

	// If all keys and values match, the maps are equal
	return true
}
