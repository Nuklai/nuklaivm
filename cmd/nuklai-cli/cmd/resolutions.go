// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/vm"

	"github.com/ava-labs/hypersdk/api/jsonrpc"
	"github.com/ava-labs/hypersdk/api/ws"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/utils"
)

// sendAndWait may not be used concurrently
func sendAndWait(
	ctx context.Context, actions []chain.Action, cli *jsonrpc.JSONRPCClient,
	bcli *vm.JSONRPCClient, ws *ws.WebSocketClient, factory chain.AuthFactory,
) (*chain.Result, ids.ID, error) {
	parser, err := bcli.Parser(ctx)
	if err != nil {
		return nil, ids.Empty, err
	}
	_, tx, _, err := cli.GenerateTransaction(ctx, parser, actions, factory)
	if err != nil {
		return nil, ids.Empty, err
	}
	if err := ws.RegisterTx(tx); err != nil {
		return nil, ids.Empty, err
	}
	var result *chain.Result
	for {
		txID, txErr, txResult, err := ws.ListenTx(ctx)
		if err != nil {
			return nil, ids.Empty, err
		}
		if txErr != nil {
			return nil, ids.Empty, txErr
		}
		if txID == tx.ID() {
			result = txResult
			break
		}
		utils.Outf("{{yellow}}skipping unexpected transaction:{{/}} %s\n", tx.ID())
	}
	status := "❌"
	if result.Success {
		status = "✅"
	}
	utils.Outf("%s {{yellow}}txID:{{/}} %s\n", status, tx.ID())

	return result, tx.ID(), nil
}

func handleTx(tx *chain.Transaction, result *chain.Result) {
	actor := tx.Auth.Actor()
	if !result.Success {
		utils.Outf(
			"%s {{yellow}}%s{{/}} {{yellow}}actor:{{/}} %s {{yellow}}error:{{/}} [%s] {{yellow}}fee (max %.2f%%):{{/}} %s %s {{yellow}}consumed:{{/}} [%s]\n",
			"❌",
			tx.ID(),
			actor,
			result.Error,
			float64(result.Fee)/float64(tx.Base.MaxFee)*100,
			utils.FormatBalance(result.Fee),
			consts.Symbol,
			result.Units,
		)
		return
	}

	for _, action := range tx.Actions {
		var summaryStr string
		switch act := action.(type) {
		case *actions.Transfer:
			summaryStr = fmt.Sprintf("assetID: %s amount: %d -> %s", act.AssetAddress, act.Value, act.To)
			if len(act.Memo) > 0 {
				summaryStr += fmt.Sprintf(" memo: %s", act.Memo)
			}
			summaryStr += "\n"
		case *actions.ContractPublish:
			summaryStr = fmt.Sprintf("contract published with txID: %s\n", tx.ID())
		case *actions.ContractDeploy:
			summaryStr = fmt.Sprintf("contractID: %s creationInfo: %s\n", string(act.ContractID), string(act.CreationInfo))
		case *actions.ContractCall:
			summaryStr = fmt.Sprintf("contractAddress: %s value: %d function: %s calldata: %s\n", act.ContractAddress, act.Value, act.Function, string(act.CallData))
		case *actions.CreateAsset:
			assetAddress := storage.AssetAddress(act.AssetType, []byte(act.Name), []byte(act.Symbol), act.Decimals, []byte(act.Metadata), actor)
			summaryStr = fmt.Sprintf("assetAddress: %s symbol: %s decimals: %d metadata: %s\n", assetAddress, act.Symbol, act.Decimals, act.Metadata)
		case *actions.UpdateAsset:
			summaryStr = fmt.Sprintf("assetAddress: %s updated\n", act.AssetAddress)
		case *actions.MintAssetFT:
			summaryStr = fmt.Sprintf("assetAddress: %s assetType: amount: %d -> %s\n", act.AssetAddress, act.Value, act.To)
		case *actions.MintAssetNFT:
			nftAddress := storage.AssetAddressNFT(act.AssetAddress, []byte(act.Metadata), act.To)
			summaryStr = fmt.Sprintf("assetAddress: %s nftAddress: %s metadata: %s -> %s\n", act.AssetAddress, nftAddress, act.Metadata, act.To)
		case *actions.BurnAssetFT:
			summaryStr = fmt.Sprintf("assetAddress: %s %d -> 🔥\n", act.AssetAddress, act.Value)
		case *actions.BurnAssetNFT:
			summaryStr = fmt.Sprintf("assetAddress: %s nftID: %s -> 🔥\n", act.AssetAddress, act.AssetNftAddress)
		case *actions.RegisterValidatorStake:
			summaryStr = fmt.Sprintf("nodeID: %s\n", act.NodeID)
		case *actions.WithdrawValidatorStake:
			summaryStr = fmt.Sprintf("nodeID: %s\n", act.NodeID)
		case *actions.ClaimValidatorStakeRewards:
			summaryStr = fmt.Sprintf("nodeID: %s\n", act.NodeID)
		case *actions.DelegateUserStake:
			summaryStr = fmt.Sprintf("nodeID: %s stakeStartBlock: %d stakeEndBlock: %d stakedAmount: %d\n", act.NodeID, act.StakeStartBlock, act.StakeEndBlock, act.StakedAmount)
		case *actions.UndelegateUserStake:
			summaryStr = fmt.Sprintf("nodeID: %s\n", act.NodeID)
		case *actions.ClaimDelegationStakeRewards:
			summaryStr = fmt.Sprintf("nodeID: %s\n", act.NodeID)
		case *actions.CreateDataset:
			summaryStr = fmt.Sprintf("datasetAddress: %s name: %s description: %s\n", act.AssetAddress, act.Name, act.Description)
		case *actions.UpdateDataset:
			summaryStr = fmt.Sprintf("datasetAddress: %s updated\n", act.DatasetAddress)
		case *actions.InitiateContributeDataset:
			summaryStr = fmt.Sprintf("datasetAddress: %s dataLocation: %s dataIdentifier: %s\n", act.DatasetAddress, act.DataLocation, act.DataIdentifier)
		case *actions.CompleteContributeDataset:
			nftAddress := codec.CreateAddress(consts.AssetFractionalTokenID, act.DatasetContributionID)
			summaryStr = fmt.Sprintf("datasetContributionID: %s datasetAddress: %s contributor: %s nftAddress: %s\n", act.DatasetContributionID, act.DatasetAddress, act.DatasetContributor, nftAddress)
		case *actions.PublishDatasetMarketplace:
			summaryStr = fmt.Sprintf("datasetAddress: %s paymentAssetAddress: %s datasetPricePerBlock: %d\n", act.DatasetAddress, act.PaymentAssetAddress, act.DatasetPricePerBlock)
		case *actions.SubscribeDatasetMarketplace:
			summaryStr = fmt.Sprintf("marketplaceAssetAddress: %s paymentAssetAddress: %s numBlocksToSubscribe: %d\n", act.MarketplaceAssetAddress, act.PaymentAssetAddress, act.NumBlocksToSubscribe)
		case *actions.ClaimMarketplacePayment:
			summaryStr = fmt.Sprintf("marketplaceAssetAddress: %s paymentAssetAddress: %s\n", act.MarketplaceAssetAddress, act.PaymentAssetAddress)
		}
		utils.Outf(
			"%s {{yellow}}%s{{/}} {{yellow}}actor:{{/}} %s {{yellow}}summary (%s):{{/}} [%s] {{yellow}}fee (max %.2f%%):{{/}} %s %s {{yellow}}consumed:{{/}} [%s]\n",
			"✅",
			tx.ID(),
			actor,
			reflect.TypeOf(action),
			summaryStr,
			float64(result.Fee)/float64(tx.Base.MaxFee)*100,
			utils.FormatBalance(result.Fee),
			consts.Symbol,
			result.Units,
		)
	}
}
