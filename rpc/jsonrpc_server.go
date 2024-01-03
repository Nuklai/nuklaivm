// Copyright (C) 2023, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emissionbalancer"
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

type BalanceArgs struct {
	Address string `json:"address"`
}

type BalanceReply struct {
	Amount uint64 `json:"amount"`
}

func (j *JSONRPCServer) Balance(req *http.Request, args *BalanceArgs, reply *BalanceReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Balance")
	defer span.End()

	addr, err := codec.ParseAddressBech32(consts.HRP, args.Address)
	if err != nil {
		return err
	}
	balance, err := j.c.GetBalanceFromState(ctx, addr)
	if err != nil {
		return err
	}
	reply.Amount = balance
	return err
}

type EmissionBalancerReply struct {
	TotalSupply     uint64 `json:"totalSupply"`
	MaxSupply       uint64 `json:"maxSupply"`
	RewardsPerBlock uint64 `json:"rewardsPerBlock"`
	Validators      map[string]*emissionbalancer.Validator
}

func (j *JSONRPCServer) EmissionBalancerInfo(req *http.Request, _ *struct{}, reply *EmissionBalancerReply) (err error) {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.EmissionBalancerInfo")
	defer span.End()

	totalSupply, maxSupply, rewardsPerBlock, validators, err := j.c.GetEmissionBalancerInfo(ctx)
	if err != nil {
		return err
	}
	reply.TotalSupply = totalSupply
	reply.MaxSupply = maxSupply
	reply.RewardsPerBlock = rewardsPerBlock
	reply.Validators = validators
	return nil
}

/* type ValidatorReply struct {
	validators map[ids.NodeID]*validators.GetValidatorOutput
} */

/* func (j *JSONRPCServer) Validators(req *http.Request, reply *ValidatorReply) error {
	utils.Outf("HERE on jsonrpc_server.go Validators\n")

	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Validators")
	defer span.End()

	validators, err := j.c.Validators(ctx)
	if err != nil {
		return err
	}
	reply.validators = validators
	return err
} */
