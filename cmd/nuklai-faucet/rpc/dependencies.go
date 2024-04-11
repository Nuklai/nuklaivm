// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
)

type Manager interface {
	GetFaucetAddress(context.Context) (codec.Address, error)
	GetChallenge(context.Context) ([]byte, uint16, error)
	SolveChallenge(context.Context, codec.Address, []byte, []byte) (ids.ID, uint64, error)
	UpdateNuklaiRPC(context.Context, string) error
}
