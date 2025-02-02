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
	"github.com/ava-labs/hypersdk/keys"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/x/contracts/runtime"

	mconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*ContractDeploy)(nil)

const MAXCREATIONSIZE = units.MiB

type ContractDeploy struct {
	ContractID   runtime.ContractID `serialize:"true" json:"contractID"`
	CreationInfo []byte             `serialize:"true" json:"creationInfo"`
	address      codec.Address
}

func (*ContractDeploy) GetTypeID() uint8 {
	return mconsts.ContractDeployID
}

func (d *ContractDeploy) StateKeys(_ codec.Address) state.Keys {
	if d.address == codec.EmptyAddress {
		d.address = storage.GetAddressForDeploy(0, d.CreationInfo)
	}
	stateKey, _ := keys.Encode(storage.AccountContractKey(d.address), 36)
	return state.Keys{
		string(stateKey): state.All,
	}
}

func (d *ContractDeploy) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) (codec.Typed, error) {
	result, err := (&storage.ContractStateManager{Mutable: mu}).
		NewAccountWithContract(ctx, d.ContractID, d.CreationInfo)
	return &ContractDeployResult{Actor: actor.String(), Receiver: "", Address: result}, err
}

func (*ContractDeploy) ComputeUnits(chain.Rules) uint64 {
	return 1
}

func (*ContractDeploy) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}

var _ chain.Marshaler = (*ContractDeploy)(nil)

func (d *ContractDeploy) Size() int {
	return codec.BytesLen(d.CreationInfo) + codec.BytesLen(d.ContractID)
}

func (d *ContractDeploy) Marshal(p *codec.Packer) {
	p.PackBytes(d.ContractID)
	p.PackBytes(d.CreationInfo)
}

func UnmarshalDeployContract(p *codec.Packer) (chain.Action, error) {
	var deployContract ContractDeploy
	p.UnpackBytes(40, true, (*[]byte)(&deployContract.ContractID))
	p.UnpackBytes(MAXCREATIONSIZE, false, &deployContract.CreationInfo)
	deployContract.address = storage.GetAddressForDeploy(0, deployContract.CreationInfo)
	if err := p.Err(); err != nil {
		return nil, err
	}

	return &deployContract, nil
}

var _ codec.Typed = (*ContractDeployResult)(nil)

type ContractDeployResult struct {
	Actor    string        `serialize:"true" json:"actor"`
	Receiver string        `serialize:"true" json:"receiver"`
	Address  codec.Address `serialize:"true" json:"address"`
}

func (*ContractDeployResult) GetTypeID() uint8 {
	return mconsts.ContractDeployID
}

func UnmarshalContractDeployResult(p *codec.Packer) (codec.Typed, error) {
	var result ContractDeployResult
	result.Actor = p.UnpackString(true)
	result.Receiver = p.UnpackString(false)
	p.UnpackAddress(&result.Address)
	return &result, p.Err()
}
