// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"fmt"
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

func CombineWithSuffix(field []byte, uniqueNum uint64, maxLength int) []byte {
	// Convert the unique number to a string and append a hyphen
	uniqueNumStr := fmt.Sprintf("-%d", uniqueNum)
	uniqueNumLen := len(uniqueNumStr)

	// Ensure there is space for the suffix
	if maxLength < uniqueNumLen {
		// If maxLength is less than the suffix length, return only the suffix (truncated if necessary)
		return []byte(uniqueNumStr)[:maxLength]
	}

	// Calculate the maximum allowable length for the name
	maxNameLen := maxLength - uniqueNumLen

	// Truncate the name if it's too long to fit with the suffix
	if len(field) > maxNameLen {
		field = field[:maxNameLen]
	}

	// Combine the truncated name with the suffix
	field = append(field, []byte(uniqueNumStr)...)

	return field
}
