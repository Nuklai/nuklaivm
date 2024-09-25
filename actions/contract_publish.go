// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"crypto/sha256"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/keys"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/x/contracts/runtime"

	mconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*ContractPublish)(nil)

const MAXCONTRACTSIZE = 2 * units.MiB

type ContractPublish struct {
	ContractBytes []byte `serialize:"true" json:"contractBytes"`
	id            runtime.ContractID
}

func (*ContractPublish) GetTypeID() uint8 {
	return mconsts.ContractPublishID
}

func (t *ContractPublish) StateKeys(_ codec.Address, _ ids.ID) state.Keys {
	if t.id == nil {
		hashedID := sha256.Sum256(t.ContractBytes)
		t.id, _ = keys.Encode(storage.ContractsKey(hashedID[:]), len(t.ContractBytes))
	}
	return state.Keys{
		string(t.id): state.Write | state.Allocate,
	}
}

func (t *ContractPublish) StateKeysMaxChunks() []uint16 {
	if chunks, ok := keys.NumChunks(t.ContractBytes); ok {
		return []uint16{chunks}
	}
	return []uint16{consts.MaxUint16}
}

func (t *ContractPublish) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	_ codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	resultBytes, err := storage.StoreContract(ctx, mu, t.ContractBytes)
	if err != nil {
		return nil, err
	}
	return &ContractPublishResult{Value: resultBytes}, nil
}

func (*ContractPublish) ComputeUnits(chain.Rules) uint64 {
	return 5
}

func (*ContractPublish) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*ContractPublish)(nil)

func (t *ContractPublish) Size() int {
	return 4 + len(t.ContractBytes)
}

func (t *ContractPublish) Marshal(p *codec.Packer) {
	p.PackBytes(t.ContractBytes)
}

func UnmarshalPublishContract(p *codec.Packer) (chain.Action, error) {
	var publishContract ContractPublish
	p.UnpackBytes(MAXCONTRACTSIZE, true, &publishContract.ContractBytes)
	if err := p.Err(); err != nil {
		return nil, err
	}

	return &publishContract, nil
}

var _ codec.Typed = (*ContractPublishResult)(nil)

type ContractPublishResult struct {
	Value []byte `serialize:"true" json:"value"`
}

func (*ContractPublishResult) GetTypeID() uint8 {
	return mconsts.ContractPublishResultID
}
