// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

// sendAndWait may not be used concurrently
func sendAndWait(
	ctx context.Context, actions []chain.Action, cli *rpc.JSONRPCClient,
	scli *rpc.WebSocketClient, tcli *nrpc.JSONRPCClient, factory chain.AuthFactory, printStatus bool,
) (ids.ID, error) {
	parser, err := tcli.Parser(ctx)
	if err != nil {
		return ids.Empty, err
	}
	_, tx, _, err := cli.GenerateTransaction(ctx, parser, actions, factory)
	if err != nil {
		return ids.Empty, err
	}

	if err := scli.RegisterTx(tx); err != nil {
		return ids.Empty, err
	}
	var res *chain.Result
	for {
		txID, dErr, result, err := scli.ListenTx(ctx)
		if dErr != nil {
			return ids.Empty, dErr
		}
		if err != nil {
			return ids.Empty, err
		}
		if txID == tx.ID() {
			res = result
			break
		}
		// TODO: don't drop these results (may be needed by a different connection)
		utils.Outf("{{yellow}}skipping unexpected transaction:{{/}} %s\n", tx.ID())
	}
	if printStatus {
		handler.Root().PrintStatus(tx.ID(), res.Success)
	}
	return tx.ID(), nil
}

func handleTx(c *nrpc.JSONRPCClient, tx *chain.Transaction, result *chain.Result) {
	actor := tx.Auth.Actor()
	if !result.Success {
		utils.Outf(
			"%s {{yellow}}%s{{/}} {{yellow}}actor:{{/}} %s {{yellow}}error:{{/}} [%s] {{yellow}}fee (max %.2f%%):{{/}} %s %s {{yellow}}consumed:{{/}} [%s]\n",
			"❌",
			tx.ID(),
			codec.MustAddressBech32(nconsts.HRP, actor),
			result.Error,
			float64(result.Fee)/float64(tx.Base.MaxFee)*100,
			utils.FormatBalance(result.Fee, nconsts.Decimals),
			nconsts.Symbol,
			cli.ParseDimensions(result.Units),
		)
		return
	}

	for i, act := range tx.Actions {
		actionID := chain.CreateActionID(tx.ID(), uint8(i))
		var summaryStr string
		switch action := act.(type) {
		case *actions.CreateAsset:
			summaryStr = fmt.Sprintf("assetID: %s symbol: %s decimals: %d metadata: %s", actionID, action.Symbol, action.Decimals, action.Metadata)
		case *actions.MintAsset:
			_, symbol, decimals, _, _, _, err := c.Asset(context.TODO(), action.Asset.String(), true)
			if err != nil {
				utils.Outf("{{red}}could not fetch asset info:{{/}} %v", err)
				return
			}
			amountStr := utils.FormatBalance(action.Value, decimals)
			summaryStr = fmt.Sprintf("%s %s -> %s", amountStr, symbol, codec.MustAddressBech32(nconsts.HRP, action.To))
		case *actions.BurnAsset:
			summaryStr = fmt.Sprintf("%d %s -> 🔥", action.Value, action.Asset)
		case *actions.Transfer:
			_, symbol, decimals, _, _, _, err := c.Asset(context.TODO(), action.Asset.String(), true)
			if err != nil {
				utils.Outf("{{red}}could not fetch asset info:{{/}} %v", err)
				return
			}
			amountStr := utils.FormatBalance(action.Value, decimals)
			summaryStr = fmt.Sprintf("%s %s -> %s", amountStr, symbol, codec.MustAddressBech32(nconsts.HRP, action.To))
			if len(action.Memo) > 0 {
				summaryStr += fmt.Sprintf(" (memo: %s)", action.Memo)
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
			summaryStr = fmt.Sprintf("nodeID: %s rewardAddress: %s", nodeID.String(), codec.MustAddressBech32(nconsts.HRP, action.RewardAddress))
		case *actions.DelegateUserStake:
			nodeID, _ := ids.ToNodeID(action.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s stakedAmount: %s rewardAddress: %s", nodeID.String(), utils.FormatBalance(action.StakedAmount, nconsts.Decimals), codec.MustAddressBech32(nconsts.HRP, action.RewardAddress))
		case *actions.ClaimDelegationStakeRewards:
			nodeID, _ := ids.ToNodeID(action.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s userStakeAddress:%s", nodeID.String(), codec.MustAddressBech32(nconsts.HRP, action.UserStakeAddress))
		case *actions.UndelegateUserStake:
			nodeID, _ := ids.ToNodeID(action.NodeID)
			summaryStr = fmt.Sprintf("nodeID: %s rewardAddress: %s", nodeID.String(), codec.MustAddressBech32(nconsts.HRP, action.RewardAddress))
		}
		utils.Outf(
			"%s {{yellow}}%s{{/}} {{yellow}}actor:{{/}} %s {{yellow}}summary (%s):{{/}} [%s] {{yellow}}fee (max %.2f%%):{{/}} %s %s {{yellow}}consumed:{{/}} [%s]\n",
			"✅",
			tx.ID(),
			codec.MustAddressBech32(nconsts.HRP, actor),
			reflect.TypeOf(act),
			summaryStr,
			float64(result.Fee)/float64(tx.Base.MaxFee)*100,
			utils.FormatBalance(result.Fee, nconsts.Decimals),
			nconsts.Symbol,
			cli.ParseDimensions(result.Units),
		)
	}
}
