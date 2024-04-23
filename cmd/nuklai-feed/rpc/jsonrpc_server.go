// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"net/http"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/cmd/nuklai-feed/manager"
	"github.com/nuklai/nuklaivm/consts"
)

type JSONRPCServer struct {
	m Manager
}

func NewJSONRPCServer(m Manager) *JSONRPCServer {
	return &JSONRPCServer{m}
}

type FeedInfoReply struct {
	Address string `json:"address"`
	Fee     uint64 `json:"fee"`
}

func (j *JSONRPCServer) FeedInfo(req *http.Request, _ *struct{}, reply *FeedInfoReply) (err error) {
	addr, fee, err := j.m.GetFeedInfo(req.Context())
	if err != nil {
		return err
	}
	reply.Address = codec.MustAddressBech32(consts.HRP, addr)
	reply.Fee = fee
	return nil
}

type FeedArgs struct {
	SubnetID string `json:"subnetID"`
	ChainID  string `json:"chainID"`
}

type FeedReply struct {
	Feed []*manager.FeedObject `json:"feed"`
}

func (j *JSONRPCServer) Feed(req *http.Request, args *FeedArgs, reply *FeedReply) (err error) {
	feed, err := j.m.GetFeed(req.Context(), args.SubnetID, args.ChainID)
	if err != nil {
		return err
	}
	reply.Feed = feed
	return nil
}

type UpdateNuklaiRPCArgs struct {
	NuklaiRPCUrl string `json:"nuklaiRPCUrl"`
}

type UpdateNuklaiRPCReply struct {
	Success bool `json:"success"`
}

func (j *JSONRPCServer) UpdateNuklaiRPC(req *http.Request, args *UpdateNuklaiRPCArgs, reply *UpdateNuklaiRPCReply) error {
	err := j.m.UpdateNuklaiRPC(req.Context(), args.NuklaiRPCUrl)
	if err != nil {
		return err
	}
	reply.Success = true
	return nil
}
