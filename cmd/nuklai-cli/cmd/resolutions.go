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
	"github.com/nuklai/nuklaivm/vm"

	"github.com/ava-labs/hypersdk/api/jsonrpc"
	"github.com/ava-labs/hypersdk/api/ws"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/utils"

	nchain "github.com/nuklai/nuklaivm/chain"
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
			utils.FormatBalance(result.Fee, consts.Decimals),
			consts.Symbol,
			result.Units,
		)
		return
	}

	for _, action := range tx.Actions {
		var summaryStr string
		switch act := action.(type) {
		case *actions.Transfer:
			summaryStr = fmt.Sprintf("assetID: %s amount: %d -> %s", act.AssetID, act.Value, act.To)
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
			summaryStr = fmt.Sprintf("assetID: %s symbol: %s decimals: %d metadata: %s\n", tx.ID(), act.Symbol, act.Decimals, act.Metadata)
		case *actions.UpdateAsset:
			summaryStr = fmt.Sprintf("assetID: %s updated\n", act.AssetID)
		case *actions.MintAssetFT:
			summaryStr = fmt.Sprintf("assetID: %s assetType: amount: %d -> %s\n", act.AssetID, act.Value, act.To)
		case *actions.MintAssetNFT:
			nftID := nchain.GenerateIDWithIndex(act.AssetID, act.UniqueID)
			summaryStr = fmt.Sprintf("assetID: %s nftID: %s uri: %s metadata: %s -> %s\n", act.AssetID, nftID, act.URI, act.Metadata, act.To)
		case *actions.BurnAssetFT:
			summaryStr = fmt.Sprintf("assetID: %s %d -> 🔥\n", act.AssetID, act.Value)
		case *actions.BurnAssetNFT:
			summaryStr = fmt.Sprintf("assetID: %s nftID: %s -> 🔥\n", act.AssetID, act.NftID)
		case *actions.CreateDataset:
			datasetID := tx.ID()
			if act.AssetID != ids.Empty {
				datasetID = act.AssetID
			}
			summaryStr = fmt.Sprintf("datasetID: %s ParentNFTID: %s name: %s description: %s\n", datasetID, nchain.GenerateIDWithIndex(datasetID, 0), act.Name, act.Description)
		case *actions.UpdateDataset:
			summaryStr = fmt.Sprintf("datasetID: %s updated\n", act.DatasetID)
		}
		utils.Outf(
			"%s {{yellow}}%s{{/}} {{yellow}}actor:{{/}} %s {{yellow}}summary (%s):{{/}} [%s] {{yellow}}fee (max %.2f%%):{{/}} %s %s {{yellow}}consumed:{{/}} [%s]\n",
			"✅",
			tx.ID(),
			actor,
			reflect.TypeOf(action),
			summaryStr,
			float64(result.Fee)/float64(tx.Base.MaxFee)*100,
			utils.FormatBalance(result.Fee, consts.Decimals),
			consts.Symbol,
			result.Units,
		)
	}
}
