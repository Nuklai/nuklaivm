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

func (*ModifyEmissionConfigParams) StateKeys(_ codec.Address, _ ids.ID) []string {
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
	_ context.Context,
	_ chain.Rules,
	_ state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	// Get the emission instance
	emissionInstance := emission.GetEmission()

	if s.MaxSupply > 0 && s.MaxSupply != emissionInstance.MaxSupply {
		if err := emissionInstance.ModifyMaxSupply(actor, s.MaxSupply); err != nil {
			return false, ModifyEmissionConfigUnits, utils.ErrBytes(err), nil, nil
		}
	}

	if s.AccountAddress != codec.EmptyAddress && s.AccountAddress != emissionInstance.EmissionAccount.Address {
		if err := emissionInstance.ModifyAccountAddress(actor, s.AccountAddress); err != nil {
			return false, ModifyEmissionConfigUnits, utils.ErrBytes(err), nil, nil
		}
	}

	baseTracker := float64(s.TrackerBaseAPR) / float64(100)
	if baseTracker > 0 && baseTracker != emissionInstance.EpochTracker.BaseAPR {
		if err := emissionInstance.ModifyBaseAPR(actor, baseTracker); err != nil {
			return false, ModifyEmissionConfigUnits, utils.ErrBytes(err), nil, nil
		}
	}
	if s.TrackerBaseValidators > 0 && s.TrackerBaseValidators != emissionInstance.EpochTracker.BaseValidators {
		if err := emissionInstance.ModifyBaseValidators(actor, s.TrackerBaseValidators); err != nil {
			return false, ModifyEmissionConfigUnits, utils.ErrBytes(err), nil, nil
		}
	}
	if s.TrackerEpochLength > 0 && s.TrackerEpochLength != emissionInstance.EpochTracker.EpochLength {
		if err := emissionInstance.ModifyEpochLength(actor, s.TrackerEpochLength); err != nil {
			return false, ModifyEmissionConfigUnits, utils.ErrBytes(err), nil, nil
		}
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
	address := config.AccountAddress[:]
	p.UnpackFixedBytes(codec.AddressLen, &address)
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

func UnmarshalModifyEmissionConfigParamsResult(b []byte) (*ModifyEmissionConfigParams, error) {
	p := codec.NewReader(b, 4*hconsts.Uint64Len+codec.AddressLen)
	var config ModifyEmissionConfigParams
	config.MaxSupply = p.UnpackUint64(false)
	config.TrackerBaseAPR = p.UnpackUint64(false)
	config.TrackerBaseValidators = p.UnpackUint64(false)
	config.TrackerEpochLength = p.UnpackUint64(false)
	address := config.AccountAddress[:]
	p.UnpackFixedBytes(codec.AddressLen, &address)
	if err := p.Err(); err != nil {
		return nil, err
	}
	return &config, nil
}

func (*ModifyEmissionConfigParams) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
