// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
)

type JSONRPCServer struct {
	c Controller
}

func NewJSONRPCServer(c Controller) *JSONRPCServer {
	return &JSONRPCServer{c}
}

type GenesisReply struct {
	Genesis *genesis.Genesis `json:"genesis"`
}

func (j *JSONRPCServer) Genesis(_ *http.Request, _ *struct{}, reply *GenesisReply) (err error) {
	reply.Genesis = j.c.Genesis()
	return nil
}

type TxArgs struct {
	TxID ids.ID `json:"txId"`
}

type TxReply struct {
	Timestamp int64            `json:"timestamp"`
	Success   bool             `json:"success"`
	Units     chain.Dimensions `json:"units"`
	Fee       uint64           `json:"fee"`
}

func (j *JSONRPCServer) Tx(req *http.Request, args *TxArgs, reply *TxReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Tx")
	defer span.End()

	found, t, success, units, fee, err := j.c.GetTransaction(ctx, args.TxID)
	if err != nil {
		return err
	}
	if !found {
		return ErrTxNotFound
	}
	reply.Timestamp = t
	reply.Success = success
	reply.Units = units
	reply.Fee = fee
	return nil
}

type AssetArgs struct {
	Asset ids.ID `json:"asset"`
}

type AssetReply struct {
	Symbol   []byte `json:"symbol"`
	Decimals uint8  `json:"decimals"`
	Metadata []byte `json:"metadata"`
	Supply   uint64 `json:"supply"`
	Owner    string `json:"owner"`
	Warp     bool   `json:"warp"`
}

func (j *JSONRPCServer) Asset(req *http.Request, args *AssetArgs, reply *AssetReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Asset")
	defer span.End()

	exists, symbol, decimals, metadata, supply, owner, warp, err := j.c.GetAssetFromState(ctx, args.Asset)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAssetNotFound
	}
	reply.Symbol = symbol
	reply.Decimals = decimals
	reply.Metadata = metadata
	reply.Supply = supply
	reply.Owner = codec.MustAddressBech32(nconsts.HRP, owner)
	reply.Warp = warp
	return err
}

type BalanceArgs struct {
	Address string `json:"address"`
	Asset   ids.ID `json:"asset"`
}

type BalanceReply struct {
	Amount uint64 `json:"amount"`
}

func (j *JSONRPCServer) Balance(req *http.Request, args *BalanceArgs, reply *BalanceReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Balance")
	defer span.End()

	addr, err := codec.ParseAddressBech32(nconsts.HRP, args.Address)
	if err != nil {
		return err
	}
	balance, err := j.c.GetBalanceFromState(ctx, addr, args.Asset)
	if err != nil {
		return err
	}
	reply.Amount = balance
	return err
}

type LoanArgs struct {
	Destination ids.ID `json:"destination"`
	Asset       ids.ID `json:"asset"`
}

type LoanReply struct {
	Amount uint64 `json:"amount"`
}

func (j *JSONRPCServer) Loan(req *http.Request, args *LoanArgs, reply *LoanReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Loan")
	defer span.End()

	amount, err := j.c.GetLoanFromState(ctx, args.Asset, args.Destination)
	if err != nil {
		return err
	}
	reply.Amount = amount
	return nil
}

type EmissionReply struct {
	TotalSupply     uint64                   `json:"totalSupply"`
	MaxSupply       uint64                   `json:"maxSupply"`
	TotalStaked     uint64                   `json:"totalStaked"`
	RewardsPerEpoch uint64                   `json:"rewardsPerEpoch"`
	EmissionAccount emission.EmissionAccount `json:"emissionAccount"`
	EpochTracker    emission.EpochTracker    `json:"epochTracker"`
}

func (j *JSONRPCServer) EmissionInfo(req *http.Request, _ *struct{}, reply *EmissionReply) (err error) {
	_, span := j.c.Tracer().Start(req.Context(), "Server.EmissionInfo")
	defer span.End()

	totalSupply, maxSupply, totalStaked, rewardsPerEpoch, emissionAccount, epochTracker, err := j.c.GetEmissionInfo()
	if err != nil {
		return err
	}
	reply.TotalSupply = totalSupply
	reply.MaxSupply = maxSupply
	reply.TotalStaked = totalStaked
	reply.RewardsPerEpoch = rewardsPerEpoch
	reply.EmissionAccount = emissionAccount
	reply.EpochTracker = epochTracker
	return nil
}

type ValidatorsReply struct {
	Validators []*emission.Validator `json:"validators"`
}

func (j *JSONRPCServer) AllValidators(req *http.Request, _ *struct{}, reply *ValidatorsReply) (err error) {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.AllValidators")
	defer span.End()

	validators, err := j.c.GetValidators(ctx, false)
	if err != nil {
		return err
	}
	reply.Validators = validators
	return nil
}

func (j *JSONRPCServer) StakedValidators(req *http.Request, _ *struct{}, reply *ValidatorsReply) (err error) {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.StakedValidators")
	defer span.End()

	validators, err := j.c.GetValidators(ctx, true)
	if err != nil {
		return err
	}
	reply.Validators = validators
	return nil
}

type ValidatorStakeArgs struct {
	NodeID ids.NodeID `json:"nodeID"`
}

type ValidatorStakeReply struct {
	StakeStartTime    uint64        `json:"stakeStartTime"`    // Start date of the stake
	StakeEndTime      uint64        `json:"stakeEndTime"`      // End date of the stake
	StakedAmount      uint64        `json:"stakedAmount"`      // Amount of NAI staked
	DelegationFeeRate uint64        `json:"delegationFeeRate"` // Delegation fee rate
	RewardAddress     codec.Address `json:"rewardAddress"`     // Address to receive rewards
	OwnerAddress      codec.Address `json:"ownerAddress"`      // Address of the owner who registered the validator
}

func (j *JSONRPCServer) ValidatorStake(req *http.Request, args *ValidatorStakeArgs, reply *ValidatorStakeReply) (err error) {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.ValidatorStake")
	defer span.End()

	exists, stakeStartTime, stakeEndTime, stakedAmount, delegationFeeRate, rewardAddress, ownerAddress, err := j.c.GetValidatorStakeFromState(ctx, args.NodeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrValidatorStakeNotFound
	}

	reply.StakeStartTime = stakeStartTime
	reply.StakeEndTime = stakeEndTime
	reply.StakedAmount = stakedAmount
	reply.DelegationFeeRate = delegationFeeRate
	reply.RewardAddress = rewardAddress
	reply.OwnerAddress = ownerAddress
	return nil
}

type UserStakeArgs struct {
	Owner  codec.Address `json:"owner"`
	NodeID ids.NodeID    `json:"nodeID"`
}

type UserStakeReply struct {
	StakeStartTime uint64        `json:"stakeStartTime"` // Start date of the stake
	StakedAmount   uint64        `json:"stakedAmount"`   // Amount of NAI staked
	RewardAddress  codec.Address `json:"rewardAddress"`  // Address to receive rewards
	OwnerAddress   codec.Address `json:"ownerAddress"`   // Address of the owner who delegated
}

func (j *JSONRPCServer) UserStake(req *http.Request, args *UserStakeArgs, reply *UserStakeReply) (err error) {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.UserStake")
	defer span.End()

	exists, stakeStartTime, stakedAmount, rewardAddress, ownerAddress, err := j.c.GetDelegatedUserStakeFromState(ctx, args.Owner, args.NodeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserStakeNotFound
	}

	reply.StakeStartTime = stakeStartTime
	reply.StakedAmount = stakedAmount
	reply.RewardAddress = rewardAddress
	reply.OwnerAddress = ownerAddress
	return nil
}
