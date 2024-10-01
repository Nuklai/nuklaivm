// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/marketplace"
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
	ActionParser    *codec.TypeParser[chain.Action]
	AuthParser      *codec.TypeParser[chain.Auth]
	OutputParser    *codec.TypeParser[codec.Typed]
	emissionTracker emission.Tracker
	marketplaceHub  marketplace.Hub
	wasmRuntime     *runtime.WasmRuntime
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
		ActionParser.Register(&actions.RegisterValidatorStake{}, actions.UnmarshalRegisterValidatorStake),
		ActionParser.Register(&actions.WithdrawValidatorStake{}, actions.UnmarshalWithdrawValidatorStake),
		ActionParser.Register(&actions.ClaimValidatorStakeRewards{}, actions.UnmarshalClaimValidatorStakeRewards),
		ActionParser.Register(&actions.DelegateUserStake{}, actions.UnmarshalDelegateUserStake),
		ActionParser.Register(&actions.UndelegateUserStake{}, actions.UnmarshalUndelegateUserStake),
		ActionParser.Register(&actions.ClaimDelegationStakeRewards{}, actions.UnmarshalClaimDelegationStakeRewards),
		ActionParser.Register(&actions.CreateDataset{}, actions.UnmarshalCreateDataset),
		ActionParser.Register(&actions.UpdateDataset{}, actions.UnmarshalUpdateDataset),
		ActionParser.Register(&actions.InitiateContributeDataset{}, actions.UnmarshalInitiateContributeDataset),
		ActionParser.Register(&actions.CompleteContributeDataset{}, actions.UnmarshalCompleteContributeDataset),
		ActionParser.Register(&actions.PublishDatasetMarketplace{}, actions.UnmarshalPublishDatasetMarketplace),
		ActionParser.Register(&actions.SubscribeDatasetMarketplace{}, actions.UnmarshalSubscribeDatasetMarketplace),
		ActionParser.Register(&actions.ClaimMarketplacePayment{}, actions.UnmarshalClaimMarketplacePayment),

		// When registering new auth, ALWAYS make sure to append at the end.
		AuthParser.Register(&auth.ED25519{}, auth.UnmarshalED25519),
		AuthParser.Register(&auth.SECP256R1{}, auth.UnmarshalSECP256R1),
		AuthParser.Register(&auth.BLS{}, auth.UnmarshalBLS),

		OutputParser.Register(&actions.TransferResult{}, actions.UnmarshalTransferResult),
		OutputParser.Register(&actions.ContractCallResult{}, nil),
		OutputParser.Register(&actions.ContractDeployResult{}, actions.UnmarshalContractDeployResult),
		OutputParser.Register(&actions.ContractPublishResult{}, nil),
		OutputParser.Register(&actions.CreateAssetResult{}, actions.UnmarshalCreateAssetResult),
		OutputParser.Register(&actions.UpdateAssetResult{}, actions.UnmarshalUpdateAssetResult),
		OutputParser.Register(&actions.MintAssetFTResult{}, actions.UnmarshalMintAssetFTResult),
		OutputParser.Register(&actions.MintAssetNFTResult{}, actions.UnmarshalMintAssetNFTResult),
		OutputParser.Register(&actions.BurnAssetFTResult{}, actions.UnmarshalBurnAssetFTResult),
		OutputParser.Register(&actions.BurnAssetNFTResult{}, actions.UnmarshalBurnAssetNFTResult),
		OutputParser.Register(&actions.RegisterValidatorStakeResult{}, actions.UnmarshalRegisterValidatorStakeResult),
		OutputParser.Register(&actions.WithdrawValidatorStakeResult{}, actions.UnmarshalWithdrawValidatorStakeResult),
		OutputParser.Register(&actions.ClaimValidatorStakeRewardsResult{}, actions.UnmarshalClaimValidatorStakeRewardsResult),
		OutputParser.Register(&actions.DelegateUserStakeResult{}, actions.UnmarshalDelegateUserStakeResult),
		OutputParser.Register(&actions.UndelegateUserStakeResult{}, actions.UnmarshalUndelegateUserStakeResult),
		OutputParser.Register(&actions.ClaimDelegationStakeRewardsResult{}, actions.UnmarshalClaimDelegationStakeRewardsResult),
		OutputParser.Register(&actions.CreateDatasetResult{}, actions.UnmarshalCreateDatasetResult),
		OutputParser.Register(&actions.UpdateDatasetResult{}, actions.UnmarshalUpdateDatasetResult),
		OutputParser.Register(&actions.InitiateContributeDatasetResult{}, actions.UnmarshalInitiateContributeDatasetResult),
		OutputParser.Register(&actions.CompleteContributeDatasetResult{}, actions.UnmarshalCompleteContributeDatasetResult),
		OutputParser.Register(&actions.PublishDatasetMarketplaceResult{}, actions.UnmarshalPublishDatasetMarketplaceResult),
		OutputParser.Register(&actions.SubscribeDatasetMarketplaceResult{}, actions.UnmarshalSubscribeDatasetMarketplaceResult),
		OutputParser.Register(&actions.ClaimMarketplacePaymentResult{}, actions.UnmarshalClaimMarketplacePaymentResult),
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
		WithEmissionBalancer(),
		WithNuklaiMarketplace(),
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
