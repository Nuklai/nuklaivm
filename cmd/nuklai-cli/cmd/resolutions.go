// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	hrpc "github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

// sendAndWait may not be used concurrently
func sendAndWait(
	ctx context.Context, warpMsg *warp.Message, action chain.Action, hcli *hrpc.JSONRPCClient,
	hws *hrpc.WebSocketClient, ncli *nrpc.JSONRPCClient, factory chain.AuthFactory, printStatus bool,
) (bool, ids.ID, error) {
	parser, err := ncli.Parser(ctx)
	if err != nil {
		return false, ids.Empty, err
	}
	_, tx, _, err := hcli.GenerateTransaction(ctx, parser, warpMsg, action, factory)
	if err != nil {
		return false, ids.Empty, err
	}

	if err := hws.RegisterTx(tx); err != nil {
		return false, ids.Empty, err
	}
	var res *chain.Result
	for {
		txID, dErr, result, err := hws.ListenTx(ctx)
		if dErr != nil {
			return false, ids.Empty, dErr
		}
		if err != nil {
			return false, ids.Empty, err
		}
		if txID == tx.ID() {
			res = result
			break
		}
		utils.Outf("{{yellow}}skipping unexpected transaction:{{/}} %s\n", tx.ID())
	}
	if printStatus {
		handler.Root().PrintStatus(tx.ID(), res.Success)
	}
	return res.Success, tx.ID(), nil
}

func handleTx(ncli *nrpc.JSONRPCClient, tx *chain.Transaction, result *chain.Result) {
	summaryStr := string(result.Output)
	actor := tx.Auth.Actor()
	status := "❌" //nolint:ineffassign // reason for ignoring
	if result.Success {
		status = "✅"
		switch action := tx.Action.(type) {
		case *actions.Transfer:
			_, symbol, decimals, _, _, _, _, err := ncli.Asset(context.TODO(), action.Asset, true)
			if err != nil {
				utils.Outf("{{red}}could not fetch asset info:{{/}} %v", err)
				return
			}
			amountStr := utils.FormatBalance(action.Value, decimals)
			summaryStr = fmt.Sprintf("%s %s -> %s", amountStr, symbol, codec.MustAddressBech32(nconsts.HRP, action.To))
			if len(action.Memo) > 0 {
				summaryStr += fmt.Sprintf(" (memo: %s)", action.Memo)
			}

		case *actions.CreateAsset:
			summaryStr = fmt.Sprintf("assetID: %s symbol: %s decimals: %d metadata: %s", tx.ID(), action.Symbol, action.Decimals, action.Metadata)
		case *actions.MintAsset:
			_, symbol, decimals, _, _, _, _, err := ncli.Asset(context.TODO(), action.Asset, true)
			if err != nil {
				utils.Outf("{{red}}could not fetch asset info:{{/}} %v", err)
				return
			}
			amountStr := utils.FormatBalance(action.Value, decimals)
			summaryStr = fmt.Sprintf("%s %s -> %s", amountStr, symbol, codec.MustAddressBech32(nconsts.HRP, action.To))
		case *actions.BurnAsset:
			summaryStr = fmt.Sprintf("%d %s -> 🔥", action.Value, action.Asset)
		case *actions.ImportAsset:
			wm := tx.WarpMessage
			signers, _ := wm.Signature.NumSigners()
			wt, _ := actions.UnmarshalWarpTransfer(wm.Payload)
			summaryStr = fmt.Sprintf("source: %s signers: %d | ", wm.SourceChainID, signers)
			if wt.Return {
				summaryStr += fmt.Sprintf("%s %s -> %s (return: %t)", utils.FormatBalance(wt.Value, wt.Decimals), wt.Symbol, codec.MustAddressBech32(nconsts.HRP, wt.To), wt.Return)
			} else {
				summaryStr += fmt.Sprintf("%s %s (new: %s, original: %s) -> %s (return: %t)", utils.FormatBalance(wt.Value, wt.Decimals), wt.Symbol, actions.ImportedAssetID(wt.Asset, wm.SourceChainID), wt.Asset, codec.MustAddressBech32(nconsts.HRP, wt.To), wt.Return)
			}
			if wt.Reward > 0 {
				summaryStr += fmt.Sprintf(" | reward: %s", utils.FormatBalance(wt.Reward, wt.Decimals))
			}
			if wt.SwapIn > 0 {
				_, outSymbol, outDecimals, _, _, _, _, err := ncli.Asset(context.TODO(), wt.AssetOut, true)
				if err != nil {
					utils.Outf("{{red}}could not fetch asset info:{{/}} %v", err)
					return
				}
				summaryStr += fmt.Sprintf(" | swap in: %s %s swap out: %s %s expiry: %d fill: %t", utils.FormatBalance(wt.SwapIn, wt.Decimals), wt.Symbol, utils.FormatBalance(wt.SwapOut, outDecimals), outSymbol, wt.SwapExpiry, action.Fill)
			}
		case *actions.ExportAsset:
			wt, _ := actions.UnmarshalWarpTransfer(result.WarpMessage.Payload)
			summaryStr = fmt.Sprintf("destination: %s | ", action.Destination)
			var outputAssetID ids.ID
			if !action.Return {
				outputAssetID = actions.ImportedAssetID(action.Asset, result.WarpMessage.SourceChainID)
				summaryStr += fmt.Sprintf("%s %s (%s) -> %s (return: %t)", utils.FormatBalance(action.Value, wt.Decimals), wt.Symbol, action.Asset, codec.MustAddressBech32(nconsts.HRP, action.To), action.Return)
			} else {
				outputAssetID = wt.Asset
				summaryStr += fmt.Sprintf("%s %s (current: %s, original: %s) -> %s (return: %t)", utils.FormatBalance(action.Value, wt.Decimals), wt.Symbol, action.Asset, wt.Asset, codec.MustAddressBech32(nconsts.HRP, action.To), action.Return)
			}
			if wt.Reward > 0 {
				summaryStr += fmt.Sprintf(" | reward: %s", utils.FormatBalance(wt.Reward, wt.Decimals))
			}
			if wt.SwapIn > 0 {
				_, outSymbol, outDecimals, _, _, _, _, err := ncli.Asset(context.TODO(), wt.AssetOut, true)
				if err != nil {
					utils.Outf("{{red}}could not fetch asset info:{{/}} %v", err)
					return
				}
				summaryStr += fmt.Sprintf(" | swap in: %s %s (%s) swap out: %s %s expiry: %d", utils.FormatBalance(wt.SwapIn, wt.Decimals), wt.Symbol, outputAssetID, utils.FormatBalance(wt.SwapOut, outDecimals), outSymbol, wt.SwapExpiry)
			}

		case *actions.RegisterValidatorStake:
			stakeInfo, _ := actions.UnmarshalValidatorStakeInfo(action.StakeInfo)
			nodeID, _ := ids.ToNodeID(stakeInfo.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s stakeStartBlock: %d stakeEndBlock: %d stakedAmount: %s delegationFeeRate: %d rewardAddress: %s", nodeID.String(), stakeInfo.StakeStartBlock, stakeInfo.StakeEndBlock, utils.FormatBalance(stakeInfo.StakedAmount, nconsts.Decimals), stakeInfo.DelegationFeeRate, codec.MustAddressBech32(nconsts.HRP, stakeInfo.RewardAddress))
		case *actions.ClaimValidatorStakeRewards:
			nodeID, _ := ids.ToNodeID(action.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s", nodeID.String())
		case *actions.WithdrawValidatorStake:
			nodeID, _ := ids.ToNodeID(action.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s", nodeID.String())
		case *actions.DelegateUserStake:
			nodeID, _ := ids.ToNodeID(action.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s stakedAmount: %s rewardAddress: %s", nodeID.String(), utils.FormatBalance(action.StakedAmount, nconsts.Decimals), codec.MustAddressBech32(nconsts.HRP, action.RewardAddress))
		case *actions.ClaimDelegationStakeRewards:
			nodeID, _ := ids.ToNodeID(action.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s userStakeAddress:%s", nodeID.String(), codec.MustAddressBech32(nconsts.HRP, action.UserStakeAddress))
		case *actions.UndelegateUserStake:
			nodeID, _ := ids.ToNodeID(action.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s", nodeID.String())
		}
		utils.Outf(
			"%s {{yellow}}%s{{/}} {{yellow}}actor:{{/}} %s {{yellow}}summary (%s):{{/}} [%s] {{yellow}}fee (max %.2f%%):{{/}} %s %s {{yellow}}consumed:{{/}} [%s]\n",
			status,
			tx.ID(),
			codec.MustAddressBech32(nconsts.HRP, actor),
			reflect.TypeOf(tx.Action),
			summaryStr,
			float64(result.Fee)/float64(tx.Base.MaxFee)*100,
			utils.FormatBalance(result.Fee, nconsts.Decimals),
			nconsts.Symbol,
			cli.ParseDimensions(result.Consumed),
		)
	}
}
