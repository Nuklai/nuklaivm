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
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/actions"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

// sendAndWait may not be used concurrently
//
//nolint:unparam
func sendAndWait(
	ctx context.Context, warpMsg *warp.Message, action chain.Action, cli *rpc.JSONRPCClient,
	scli *rpc.WebSocketClient, ncli *nrpc.JSONRPCClient, factory chain.AuthFactory, printStatus bool,
) (bool, ids.ID, error) {
	parser, err := ncli.Parser(ctx)
	if err != nil {
		return false, ids.Empty, err
	}
	_, tx, _, err := cli.GenerateTransaction(ctx, parser, warpMsg, action, factory)
	if err != nil {
		return false, ids.Empty, err
	}

	if err := scli.RegisterTx(tx); err != nil {
		return false, ids.Empty, err
	}
	var res *chain.Result
	for {
		txID, dErr, result, err := scli.ListenTx(ctx)
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

func handleTx(c *nrpc.JSONRPCClient, tx *chain.Transaction, result *chain.Result) {
	summaryStr := string(result.Output)
	actor := tx.Auth.Actor()
	status := "❌"
	if result.Success {
		status = "✅"
		switch action := tx.Action.(type) { //nolint:gocritic
		case *actions.Transfer:
			_, symbol, decimals, _, _, _, _, err := c.Asset(context.TODO(), action.Asset, true)
			if err != nil {
				utils.Outf("{{red}}could not fetch asset info:{{/}} %v", err)
				return
			}
			amountStr := utils.FormatBalance(action.Value, decimals)
			summaryStr = fmt.Sprintf("%s %s -> %s", amountStr, symbol, codec.MustAddressBech32(nconsts.HRP, action.To))
			if len(action.Memo) > 0 {
				summaryStr += fmt.Sprintf(" (memo: %s)", action.Memo)
			}
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
