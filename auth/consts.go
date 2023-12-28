// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package auth

import (
	"github.com/ava-labs/hypersdk/vm"
	"github.com/nuklai/nuklaivm/consts"
)

// Note: Registry will error during initialization if a duplicate ID is assigned. We explicitly assign IDs to avoid accidental remapping.
const (
	ed25519ID uint8 = 0
)

func Engines() map[uint8]vm.AuthEngine {
	return map[uint8]vm.AuthEngine{
		// Only ed25519 batch verification is supported
		consts.ED25519ID: &ED25519AuthEngine{},
	}
}
