// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"math"
	"strconv"
)

func ParseBalance(bal string, decimals uint8) (uint64, error) {
	f, err := strconv.ParseFloat(bal, 64)
	if err != nil {
		return 0, err
	}
	return uint64(f * math.Pow10(int(decimals))), nil
}

func FormatBalance(bal uint64, decimals uint8) string {
	return strconv.FormatFloat(float64(bal)/math.Pow10(int(decimals)), 'f', int(decimals), 64)
}

// Function to combine the prefix with the name byte slice
func CombineWithPrefix(prefix, name []byte, maxLength int) []byte {
	prefixLen := len(prefix)

	// Calculate the maximum allowable length for the name
	maxNameLen := maxLength - prefixLen

	// Truncate the name if it's too long
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}

	// Combine the prefix with the (potentially truncated) name
	prefix = append(prefix, name...)

	return prefix
}
