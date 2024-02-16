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
	"github.com/nuklai/nuklaivm/storage"

	nconsts "github.com/nuklai/nuklaivm/consts"
)

var _ chain.Action = (*StakeValidator)(nil)

type StakeValidator struct {
	NodeID       []byte `json:"nodeID"`
	StakedAmount uint64 `json:"stakedAmount"`
	EndLockUp    uint64 `json:"endLockUp"`
}

func (*StakeValidator) GetTypeID() uint8 {
	return nconsts.StakeValidatorID
}

func (*StakeValidator) StateKeys(actor codec.Address, txID ids.ID) []string {
	return []string{
		string(storage.BalanceKey(actor, ids.Empty)),
		string(storage.StakeKey(txID)),
	}
}

func (*StakeValidator) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.BalanceChunks, storage.StakeChunks}
}

func (*StakeValidator) OutputsWarpMessage() bool {
	return false
}

func (s *StakeValidator) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	txID ids.ID,
	_ bool,
) (bool, uint64, []byte, *warp.UnsignedMessage, error) {
	if s.StakedAmount == 0 {
		return false, StakeValidatorComputeUnits, OutputStakedAmountInvalid, nil, nil
	}
	nodeID, err := ids.ToNodeID(s.NodeID)
	if err != nil {
		return false, StakeValidatorComputeUnits, OutputInvalidNodeID, nil, nil
	}
	if err := storage.SubBalance(ctx, mu, actor, ids.Empty, s.StakedAmount); err != nil {
		return false, StakeValidatorComputeUnits, utils.ErrBytes(err), nil, nil
	}
	if err := storage.SetStake(ctx, mu, txID, nodeID, s.StakedAmount, s.EndLockUp, actor); err != nil {
		return false, StakeValidatorComputeUnits, utils.ErrBytes(err), nil, nil
	}
	return true, StakeValidatorComputeUnits, nil, nil, nil
}

func (*StakeValidator) MaxComputeUnits(chain.Rules) uint64 {
	return StakeValidatorComputeUnits
}

func (*StakeValidator) Size() int {
	return hconsts.NodeIDLen + 2*hconsts.Uint64Len
}

func (s *StakeValidator) Marshal(p *codec.Packer) {
	p.PackBytes(s.NodeID)
	p.PackUint64(s.StakedAmount)
	p.PackUint64(s.EndLockUp)
}

func UnmarshalStakeValidator(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var stake StakeValidator
	p.UnpackBytes(hconsts.NodeIDLen, false, &stake.NodeID)
	stake.StakedAmount = p.UnpackUint64(true)
	stake.EndLockUp = p.UnpackUint64(true)
	return &stake, p.Err()
}

func (*StakeValidator) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
