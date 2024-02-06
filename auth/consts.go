// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package auth

import (
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/crypto/secp256r1"
	"github.com/ava-labs/hypersdk/vm"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	ED25519ComputeUnits = 5
	ED25519Size         = ed25519.PublicKeyLen + ed25519.SignatureLen

	SECP256R1ComputeUnits = 10 // can't be batched like ed25519
	SECP256R1Size         = secp256r1.PublicKeyLen + secp256r1.SignatureLen

	BLSComputeUnits = 10
	BLSSize         = bls.PublicKeyLen + bls.SignatureLen
)

func Engines() map[uint8]vm.AuthEngine {
	return map[uint8]vm.AuthEngine{
		// Only ed25519 batch verification is supported
		// We are not adding secp256r1 because it cannot be batched
		nconsts.ED25519ID: &ED25519AuthEngine{},
	}
}
