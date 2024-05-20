// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	hconsts "github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/emission"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*ModifyEmissionConfigParams)(nil)

type ModifyEmissionConfigParams struct {
	MaxSupply             uint64        `json:"maxSupply"`             // Emission new MaxSupply
	TrackerBaseAPR        uint64        `json:"trackerBaseAPR"`        // Emission new EpochTracker.BaseAPR * 100
	TrackerBaseValidators uint64        `json:"trackerBaseValidators"` // Emission new EpochTracker.BaseValidators
	TrackerEpochLength    uint64        `json:"trackerEpochLength"`    // Emission new EpochTracker.EpochLength
	AccountAddress        codec.Address `json:"accountAddress"`        // Emission new EmissionAccount.Address
}

func (*ModifyEmissionConfigParams) GetTypeID() uint8 {
	return nconsts.ModifyEmissionConfigParamsID
}

func (s *ModifyEmissionConfigParams) StateKeys(actor codec.Address, _ ids.ID) []string {
	// TODO: How to better handle a case where the NodeID is invalid?
	return []string{}
}

func (*ModifyEmissionConfigParams) StateKeysMaxChunks() []uint16 {
	return []uint16{}
}

func (*ModifyEmissionConfigParams) OutputsWarpMessage() bool {
	return false
}

func (s *ModifyEmissionConfigParams) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	// Get the emission instance
	emissionInstance := emission.GetEmission()

	if s.MaxSupply > 0 && s.MaxSupply != emissionInstance.MaxSupply {
		emissionInstance.ModifyMaxSupply(s.MaxSupply)
	}

	if s.AccountAddress != codec.EmptyAddress && s.AccountAddress != emissionInstance.EmissionAccount.Address {
		emissionInstance.ModifyAccountAddress(s.AccountAddress)
	}

	baseTracker := float64(s.TrackerBaseAPR / 100)
	if baseTracker > 0 && baseTracker != emissionInstance.EpochTracker.BaseAPR {
		emissionInstance.ModifyBaseAPR(baseTracker)
	}

	if s.TrackerBaseValidators > 0 && s.TrackerBaseValidators != emissionInstance.EpochTracker.BaseValidators {
		emissionInstance.ModifyBaseValidators(s.TrackerBaseValidators)
	}

	if s.TrackerEpochLength > 0 && s.TrackerEpochLength != emissionInstance.EpochTracker.EpochLength {
		emissionInstance.ModifyEpochLength(s.TrackerEpochLength)
	}

	sr := &ModifyEmissionConfigParamsResult{s}
	output, err := sr.Marshal()
	if err != nil {
		return false, ModifyEmissionConfigUnits, utils.ErrBytes(err), nil, nil
	}
	return true, ModifyEmissionConfigUnits, output, nil, nil
}

func (*ModifyEmissionConfigParams) MaxComputeUnits(chain.Rules) uint64 {
	return ModifyEmissionConfigUnits
}

func (*ModifyEmissionConfigParams) Size() int {
	return 4*hconsts.Uint64Len + codec.AddressLen
}

func (s *ModifyEmissionConfigParams) Marshal(p *codec.Packer) {
	p.PackUint64(s.MaxSupply)
	p.PackUint64(s.TrackerBaseAPR)
	p.PackUint64(s.TrackerBaseValidators)
	p.PackUint64(s.TrackerEpochLength)
	p.PackAddress(s.AccountAddress)
}

func UnmarshalModifyEmissionConfigParams(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var config ModifyEmissionConfigParams
	config.MaxSupply = p.UnpackUint64(false)
	config.TrackerBaseAPR = p.UnpackUint64(false)
	config.TrackerBaseValidators = p.UnpackUint64(false)
	config.TrackerEpochLength = p.UnpackUint64(false)
	p.UnpackAddress(&config.AccountAddress)
	if err := p.Err(); err != nil {
		return nil, err
	}
	return &config, nil
}

type ModifyEmissionConfigParamsResult struct {
	*ModifyEmissionConfigParams
}

func (s *ModifyEmissionConfigParamsResult) Marshal() ([]byte, error) {
	p := codec.NewWriter(4*hconsts.Uint64Len+codec.AddressLen, 4*hconsts.Uint64Len+codec.AddressLen)
	p.PackUint64(s.MaxSupply)
	p.PackUint64(s.TrackerBaseAPR)
	p.PackUint64(s.TrackerBaseValidators)
	p.PackUint64(s.TrackerEpochLength)
	p.PackAddress(s.AccountAddress)
	return p.Bytes(), p.Err()
}

func UnmarshalModifyEmissionConfigParamsResult(b []byte) (*ModifyEmissionConfigParamsResult, error) {
	var config ModifyEmissionConfigParamsResult
	p := codec.NewReader(b, 4*hconsts.Uint64Len+codec.AddressLen)
	config.MaxSupply = p.UnpackUint64(false)
	config.TrackerBaseAPR = p.UnpackUint64(false)
	config.TrackerBaseValidators = p.UnpackUint64(false)
	config.TrackerEpochLength = p.UnpackUint64(false)
	p.UnpackAddress(&config.AccountAddress)
	if err := p.Err(); err != nil {
		return nil, err
	}
	return &config, nil

}

func (*ModifyEmissionConfigParams) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
