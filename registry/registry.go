// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package registry

import (
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

// Setup types
func init() {
	nconsts.ActionRegistry = codec.NewTypeParser[chain.Action]()
	nconsts.AuthRegistry = codec.NewTypeParser[chain.Auth]()

	errs := &wrappers.Errs{}
	errs.Add(
		// When registering new actions, ALWAYS make sure to append at the end.
		nconsts.ActionRegistry.Register((&actions.Transfer{}).GetTypeID(), actions.UnmarshalTransfer, false),

		nconsts.ActionRegistry.Register((&actions.CreateAsset{}).GetTypeID(), actions.UnmarshalCreateAsset, false),
		nconsts.ActionRegistry.Register((&actions.UpdateAsset{}).GetTypeID(), actions.UnmarshalUpdateAsset, false),
		nconsts.ActionRegistry.Register((&actions.MintAssetFT{}).GetTypeID(), actions.UnmarshalMintAsset, false),
		nconsts.ActionRegistry.Register((&actions.MintAssetNFT{}).GetTypeID(), actions.UnmarshalMintAssetNFT, false),
		nconsts.ActionRegistry.Register((&actions.BurnAssetFT{}).GetTypeID(), actions.UnmarshalBurnAssetFT, false),
		nconsts.ActionRegistry.Register((&actions.BurnAssetNFT{}).GetTypeID(), actions.UnmarshalBurnAssetNFT, false),

		nconsts.ActionRegistry.Register((&actions.RegisterValidatorStake{}).GetTypeID(), actions.UnmarshalRegisterValidatorStake, false),
		nconsts.ActionRegistry.Register((&actions.ClaimValidatorStakeRewards{}).GetTypeID(), actions.UnmarshalClaimValidatorStakeRewards, false),
		nconsts.ActionRegistry.Register((&actions.WithdrawValidatorStake{}).GetTypeID(), actions.UnmarshalWithdrawValidatorStake, false),
		nconsts.ActionRegistry.Register((&actions.DelegateUserStake{}).GetTypeID(), actions.UnmarshalDelegateUserStake, false),
		nconsts.ActionRegistry.Register((&actions.ClaimDelegationStakeRewards{}).GetTypeID(), actions.UnmarshalClaimDelegationStakeRewards, false),
		nconsts.ActionRegistry.Register((&actions.UndelegateUserStake{}).GetTypeID(), actions.UnmarshalUndelegateUserStake, false),

		nconsts.ActionRegistry.Register((&actions.CreateDataset{}).GetTypeID(), actions.UnmarshalCreateDataset, false),
		nconsts.ActionRegistry.Register((&actions.UpdateDataset{}).GetTypeID(), actions.UnmarshalUpdateDataset, false),
		nconsts.ActionRegistry.Register((&actions.InitiateContributeDataset{}).GetTypeID(), actions.UnmarshalInitiateContributeDataset, false),
		nconsts.ActionRegistry.Register((&actions.CompleteContributeDataset{}).GetTypeID(), actions.UnmarshalCompleteContributeDataset, false),
		nconsts.ActionRegistry.Register((&actions.PublishDatasetMarketplace{}).GetTypeID(), actions.UnmarshalPublishDatasetMarketplace, false),
		nconsts.ActionRegistry.Register((&actions.SubscribeDatasetMarketplace{}).GetTypeID(), actions.UnmarshalSubscribeDatasetMarketplace, false),
		nconsts.ActionRegistry.Register((&actions.ClaimMarketplacePayment{}).GetTypeID(), actions.UnmarshalClaimMarketplacePayment, false),

		// When registering new auth, ALWAYS make sure to append at the end.
		nconsts.AuthRegistry.Register((&auth.ED25519{}).GetTypeID(), auth.UnmarshalED25519, false),
		nconsts.AuthRegistry.Register((&auth.SECP256R1{}).GetTypeID(), auth.UnmarshalSECP256R1, false),
		nconsts.AuthRegistry.Register((&auth.BLS{}).GetTypeID(), auth.UnmarshalBLS, false),
	)
	if errs.Errored() {
		panic(errs.Err)
	}
}
