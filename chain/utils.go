// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package chain

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"
)

// generateRandomID creates a random [ids.ID] for use in generating random addresses.
func GenerateRandomID() (ids.ID, error) {
	// Create a byte slice with the length of ids.ID
	randomBytes := make([]byte, ids.IDLen)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return ids.Empty, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return ids.ToID(randomBytes)
}

func GenerateIDWithIndex(id ids.ID, i uint64) ids.ID {
	actionBytes := make([]byte, ids.IDLen+consts.Uint64Len)
	copy(actionBytes, id[:])
	binary.BigEndian.PutUint64(actionBytes[ids.IDLen:], i)
	return utils.ToID(actionBytes)
}

func GenerateIDWithAddress(id ids.ID, addr codec.Address) ids.ID {
	actionBytes := make([]byte, ids.IDLen+codec.AddressLen)
	copy(actionBytes, id[:])
	copy(actionBytes[ids.IDLen:], addr[:])
	return utils.ToID(actionBytes)
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

// Function to convert a JSON string to map[string]string
func JSONToMap(jsonStr string) (map[string]string, error) {
	// Create a map to hold the unmarshaled JSON
	var result map[string]string

	// Unmarshal the JSON string into the map
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}

	// Return the map and any error encountered
	return result, nil
}
