// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
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

// GenerateIDWithString creates a new ID based on the hash of the provided string.
func GenerateIDWithString(str string) ids.ID {
	assetID, err := ids.FromString(str)
	if err != nil {
		// Create a SHA256 hash of the input string
		hash := sha256.Sum256([]byte(str))
		assetID = utils.ToID(hash[:ids.IDLen])
	}
	return assetID
}
