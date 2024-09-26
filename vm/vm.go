// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/api/indexer"
	"github.com/ava-labs/hypersdk/api/jsonrpc"
	"github.com/ava-labs/hypersdk/api/ws"
	"github.com/ava-labs/hypersdk/auth"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/extension/externalsubscriber"
	"github.com/ava-labs/hypersdk/vm"
	"github.com/ava-labs/hypersdk/x/contracts/runtime"

	staterpc "github.com/ava-labs/hypersdk/api/state"
)

var (
	ActionParser *codec.TypeParser[chain.Action]
	AuthParser   *codec.TypeParser[chain.Auth]
	OutputParser *codec.TypeParser[codec.Typed]
	wasmRuntime  *runtime.WasmRuntime
)

// Setup types
func init() {
	ActionParser = codec.NewTypeParser[chain.Action]()
	AuthParser = codec.NewTypeParser[chain.Auth]()
	OutputParser = codec.NewTypeParser[codec.Typed]()

	errs := &wrappers.Errs{}
	errs.Add(
		// When registering new actions, ALWAYS make sure to append at the end.
		// Pass nil as second argument if manual marshalling isn't needed (if in doubt, you probably don't)
		ActionParser.Register(&actions.Transfer{}, actions.UnmarshalTransfer),
		ActionParser.Register(&actions.ContractCall{}, actions.UnmarshalCallContract(wasmRuntime)),
		ActionParser.Register(&actions.ContractPublish{}, actions.UnmarshalPublishContract),
		ActionParser.Register(&actions.ContractDeploy{}, actions.UnmarshalDeployContract),
		ActionParser.Register(&actions.CreateAsset{}, actions.UnmarshalCreateAsset),
		ActionParser.Register(&actions.UpdateAsset{}, actions.UnmarshalUpdateAsset),
		ActionParser.Register(&actions.MintAssetFT{}, actions.UnmarshalMintAssetFT),
		ActionParser.Register(&actions.MintAssetNFT{}, actions.UnmarshalMintAssetNFT),
		ActionParser.Register(&actions.BurnAssetFT{}, actions.UnmarshalBurnAssetFT),
		ActionParser.Register(&actions.BurnAssetNFT{}, actions.UnmarshalBurnAssetNFT),
		ActionParser.Register(&actions.CreateDataset{}, actions.UnmarshalCreateDataset),
		ActionParser.Register(&actions.UpdateDataset{}, actions.UnmarshalUpdateDataset),

		// When registering new auth, ALWAYS make sure to append at the end.
		AuthParser.Register(&auth.ED25519{}, auth.UnmarshalED25519),
		AuthParser.Register(&auth.SECP256R1{}, auth.UnmarshalSECP256R1),
		AuthParser.Register(&auth.BLS{}, auth.UnmarshalBLS),

		OutputParser.Register(&actions.TransferResult{}, actions.UnmarshalTransferResult),
		OutputParser.Register(&actions.ContractCallResult{}, nil),
		OutputParser.Register(&actions.ContractDeployResult{}, nil),
		OutputParser.Register(&actions.ContractPublishResult{}, nil),
		OutputParser.Register(&actions.CreateAssetResult{}, nil),
		OutputParser.Register(&actions.UpdateAssetResult{}, nil),
		OutputParser.Register(&actions.MintAssetFTResult{}, nil),
		OutputParser.Register(&actions.MintAssetNFTResult{}, nil),
		OutputParser.Register(&actions.BurnAssetFTResult{}, nil),
		OutputParser.Register(&actions.BurnAssetNFTResult{}, nil),
		OutputParser.Register(&actions.CreateDatasetResult{}, nil),
		OutputParser.Register(&actions.UpdateDatasetResult{}, nil),
	)
	if errs.Errored() {
		panic(errs.Err)
	}
}

// New returns a VM with the indexer, websocket, rpc, and external subscriber apis enabled.
func New(options ...vm.Option) (*vm.VM, error) {
	opts := append([]vm.Option{
		indexer.With(),
		ws.With(),
		jsonrpc.With(),
		With(), // Add Controller API
		externalsubscriber.With(),
		staterpc.With(),
	}, options...)

	return NewWithOptions(opts...)
}

// NewWithOptions returns a VM with the specified options
func NewWithOptions(options ...vm.Option) (*vm.VM, error) {
	opts := append([]vm.Option{
		WithRuntime(),
	}, options...)
	return vm.New(
		consts.Version,
		genesis.GenesisFactory{},
		&storage.StateManager{},
		ActionParser,
		AuthParser,
		OutputParser,
		auth.Engines(),
		opts...,
	)
}
