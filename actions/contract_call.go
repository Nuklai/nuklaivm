// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/x/contracts/runtime"

	mconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*ContractCall)(nil)

const MaxCallDataSize = units.MiB

type StateKeyPermission struct {
	Key        string
	Permission state.Permissions
}

type ContractCall struct {
	// contract is the address of the contract to be called
	ContractAddress codec.Address `json:"contractAddress"`

	// Amount are transferred to [To].
	Value uint64 `json:"value"`

	// Function is the name of the function to call on the contract.
	Function string `json:"function"`

	// CallData are the serialized parameters to be passed to the contract.
	CallData []byte `json:"calldata"`

	SpecifiedStateKeys []StateKeyPermission `json:"statekeys"`

	Fuel uint64 `json:"fuel"`

	r *runtime.WasmRuntime
}

func (*ContractCall) GetTypeID() uint8 {
	return mconsts.ContractCallID
}

func (t *ContractCall) StateKeys(_ codec.Address, _ ids.ID) state.Keys {
	result := state.Keys{}
	for _, stateKeyPermission := range t.SpecifiedStateKeys {
		result.Add(stateKeyPermission.Key, stateKeyPermission.Permission)
	}
	return result
}

func (t *ContractCall) StateKeysMaxChunks() []uint16 {
	result := make([]uint16, 0, len(t.SpecifiedStateKeys))
	for range t.SpecifiedStateKeys {
		result = append(result, 1)
	}
	return result
}

func (t *ContractCall) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	timestamp int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	resutBytes, err := t.r.CallContract(ctx, &runtime.CallInfo{
		Contract:     t.ContractAddress,
		Actor:        actor,
		State:        &storage.ContractStateManager{Mutable: mu},
		FunctionName: t.Function,
		Params:       t.CallData,
		Timestamp:    uint64(timestamp),
		Fuel:         t.Fuel,
		Value:        t.Value,
	})
	if err != nil {
		return nil, err
	}
	return &ContractCallResult{Value: resutBytes}, nil
}

func (t *ContractCall) ComputeUnits(chain.Rules) uint64 {
	return t.Fuel / 1000
}

func (*ContractCall) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*ContractCall)(nil)

func (t *ContractCall) Size() int {
	return codec.AddressLen + 2*consts.Uint64Len + len(t.CallData) + len(t.Function) + len(t.SpecifiedStateKeys)
}

func (t *ContractCall) Marshal(p *codec.Packer) {
	p.PackUint64(t.Value)
	p.PackUint64(t.Fuel)
	p.PackAddress(t.ContractAddress)
	p.PackString(t.Function)
	p.PackBytes(t.CallData)
	p.PackInt(uint32(len(t.SpecifiedStateKeys)))

	for _, stateKeyPermission := range t.SpecifiedStateKeys {
		p.PackString(stateKeyPermission.Key)
		p.PackByte(byte(stateKeyPermission.Permission))
	}
}

func UnmarshalCallContract(r *runtime.WasmRuntime) func(p *codec.Packer) (chain.Action, error) {
	return func(p *codec.Packer) (chain.Action, error) {
		callContract := ContractCall{r: r}
		callContract.Value = p.UnpackUint64(false)
		callContract.Fuel = p.UnpackUint64(true)
		p.UnpackAddress(&callContract.ContractAddress) // we do not verify the typeID is valid
		callContract.Function = p.UnpackString(true)
		p.UnpackBytes(MaxCallDataSize, false, &callContract.CallData)
		if err := p.Err(); err != nil {
			return nil, err
		}
		count := int(p.UnpackInt(true))
		callContract.SpecifiedStateKeys = make([]StateKeyPermission, count)
		for i := 0; i < count; i++ {
			key := p.UnpackString(true)
			value := p.UnpackByte()
			callContract.SpecifiedStateKeys[i] = StateKeyPermission{Key: key, Permission: state.Permissions(value)}
		}
		return &callContract, nil
	}
}

var _ codec.Typed = (*ContractCallResult)(nil)

type ContractCallResult struct {
	Value []byte `serialize:"true" json:"value"`
}

func (*ContractCallResult) GetTypeID() uint8 {
	return mconsts.ContractCallResultID
}
