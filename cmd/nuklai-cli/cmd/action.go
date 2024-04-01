// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/spf13/cobra"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	hyperCli "github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	hconsts "github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/pubsub"
	hrpc "github.com/ava-labs/hypersdk/rpc"
	hutils "github.com/ava-labs/hypersdk/utils"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	frpc "github.com/nuklai/nuklaivm/cmd/nuklai-faucet/rpc"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

var actionCmd = &cobra.Command{
	Use: "action",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var fundFaucetCmd = &cobra.Command{
	Use: "fund-faucet",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()

		// Get faucet
		faucetURI, err := handler.Root().PromptString("faucet URI", 0, consts.MaxInt)
		if err != nil {
			return err
		}
		fcli := frpc.NewJSONRPCClient(faucetURI)
		faucetAddress, err := fcli.FaucetAddress(ctx)
		if err != nil {
			return err
		}

		// Get clients
		_, priv, factory, cli, scli, tcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get balance
		_, decimals, balance, _, err := handler.GetAssetInfo(ctx, tcli, priv.Address, ids.Empty, true)
		if balance == 0 || err != nil {
			return err
		}

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", decimals, balance, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		addr, err := codec.ParseAddressBech32(nconsts.HRP, faucetAddress)
		if err != nil {
			return err
		}
		if _, _, err = sendAndWait(ctx, nil, &actions.Transfer{
			To:    addr,
			Asset: ids.Empty,
			Value: amount,
		}, cli, scli, tcli, factory, true); err != nil {
			return err
		}
		hutils.Outf("{{green}}funded faucet:{{/}} %s\n", faucetAddress)
		return nil
	},
}

var transferCmd = &cobra.Command{
	Use: "transfer",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to send
		assetID, err := handler.Root().PromptAsset("assetID", true)
		if err != nil {
			return err
		}
		_, decimals, balance, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetID, true)
		if balance == 0 || err != nil {
			return err
		}

		// Select recipient
		recipient, err := handler.Root().PromptAddress("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", decimals, balance, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.Transfer{
			To:    recipient,
			Asset: assetID,
			Value: amount,
		}, hcli, hws, ncli, factory, true)
		return err
	},
}

var createAssetCmd = &cobra.Command{
	Use: "create-asset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Add symbol to token
		symbol, err := handler.Root().PromptString("symbol", 1, actions.MaxSymbolSize)
		if err != nil {
			return err
		}

		// Add decimal to token
		decimals, err := handler.Root().PromptInt("decimals", actions.MaxDecimals)
		if err != nil {
			return err
		}

		// Add metadata to token
		metadata, err := handler.Root().PromptString("metadata", 1, actions.MaxMetadataSize)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.CreateAsset{
			Symbol:   []byte(symbol),
			Decimals: uint8(decimals), // already constrain above to prevent overflow
			Metadata: []byte(metadata),
		}, hcli, hws, ncli, factory, true)
		return err
	},
}

var mintAssetCmd = &cobra.Command{
	Use: "mint-asset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to mint
		assetID, err := handler.Root().PromptAsset("assetID", false)
		if err != nil {
			return err
		}
		exists, symbol, decimals, metadata, supply, owner, warp, err := ncli.Asset(ctx, assetID, false)
		if err != nil {
			return err
		}
		if !exists {
			hutils.Outf("{{red}}%s does not exist{{/}}\n", assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		if warp {
			hutils.Outf("{{red}}cannot mint a warped asset{{/}}\n", assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		if owner != codec.MustAddressBech32(nconsts.HRP, priv.Address) {
			hutils.Outf("{{red}}%s is the owner of %s, you are not{{/}}\n", owner, assetID)
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
		hutils.Outf(
			"{{yellow}}symbol:{{/}} %s {{yellow}}decimals:{{/}} %d {{yellow}}metadata:{{/}} %s {{yellow}}supply:{{/}} %d\n",
			string(symbol),
			decimals,
			string(metadata),
			supply,
		)

		// Select recipient
		recipient, err := handler.Root().PromptAddress("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", decimals, hconsts.MaxUint64-supply, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.MintAsset{
			Asset: assetID,
			To:    recipient,
			Value: amount,
		}, hcli, hws, ncli, factory, true)
		return err
	},
}

func performImport(
	ctx context.Context,
	hscli *hrpc.JSONRPCClient,
	hdcli *hrpc.JSONRPCClient,
	hws *hrpc.WebSocketClient,
	ncli *nrpc.JSONRPCClient,
	exportTxID ids.ID,
	factory chain.AuthFactory,
) error {
	// Select TxID (if not provided)
	var err error
	if exportTxID == ids.Empty {
		exportTxID, err = handler.Root().PromptID("export txID")
		if err != nil {
			return err
		}
	}

	// Generate warp signature (as long as >= 80% stake)
	var (
		msg                     *warp.Message
		subnetWeight, sigWeight uint64
	)
	for ctx.Err() == nil {
		msg, subnetWeight, sigWeight, err = hscli.GenerateAggregateWarpSignature(ctx, exportTxID)
		if sigWeight >= (subnetWeight*4)/5 && err == nil {
			break
		}
		if err == nil {
			hutils.Outf(
				"{{yellow}}waiting for signature weight:{{/}} %d {{yellow}}observed:{{/}} %d\n",
				subnetWeight,
				sigWeight,
			)
		} else {
			hutils.Outf("{{red}}encountered error:{{/}} %v\n", err)
		}
		cont, err := handler.Root().PromptBool("try again")
		if err != nil {
			return err
		}
		if !cont {
			hutils.Outf("{{red}}exiting...{{/}}\n")
			return nil
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	wt, err := actions.UnmarshalWarpTransfer(msg.UnsignedMessage.Payload)
	if err != nil {
		return err
	}
	outputAssetID := wt.Asset
	if !wt.Return {
		outputAssetID = actions.ImportedAssetID(wt.Asset, msg.SourceChainID)
	}
	hutils.Outf(
		"%s {{yellow}}to:{{/}} %s {{yellow}}source assetID:{{/}} %s {{yellow}}source symbol:{{/}} %s {{yellow}}output assetID:{{/}} %s {{yellow}}value:{{/}} %s {{yellow}}reward:{{/}} %s {{yellow}}return:{{/}} %t\n",
		hutils.ToID(
			msg.UnsignedMessage.Payload,
		),
		codec.MustAddressBech32(nconsts.HRP, wt.To),
		wt.Asset,
		wt.Symbol,
		outputAssetID,
		hutils.FormatBalance(wt.Value, wt.Decimals),
		hutils.FormatBalance(wt.Reward, wt.Decimals),
		wt.Return,
	)
	if wt.SwapIn > 0 {
		_, outSymbol, outDecimals, _, _, _, _, err := ncli.Asset(ctx, wt.AssetOut, false)
		if err != nil {
			return err
		}
		hutils.Outf(
			"{{yellow}}asset in:{{/}} %s {{yellow}}swap in:{{/}} %s {{yellow}}asset out:{{/}} %s {{yellow}}symbol out:{{/}} %s {{yellow}}swap out:{{/}} %s {{yellow}}swap expiry:{{/}} %d\n",
			outputAssetID,
			hutils.FormatBalance(wt.SwapIn, wt.Decimals),
			wt.AssetOut,
			outSymbol,
			hutils.FormatBalance(wt.SwapOut, outDecimals),
			wt.SwapExpiry,
		)
	}

	// Select fill
	var fill bool
	if wt.SwapIn > 0 {
		fill, err = handler.Root().PromptBool("fill")
		if err != nil {
			return err
		}
	}
	if !fill && wt.SwapExpiry > time.Now().UnixMilli() {
		return ErrMustFill
	}

	// Generate transaction
	_, _, err = sendAndWait(ctx, msg, &actions.ImportAsset{
		Fill: fill,
	}, hdcli, hws, ncli, factory, true)
	return err
}

var importAssetCmd = &cobra.Command{
	Use: "import-asset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()

		currentChainID, _, factory, hdcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select source
		_, uris, err := handler.Root().PromptChain("sourceChainID", set.Of(currentChainID))
		if err != nil {
			return err
		}
		hscli := hrpc.NewJSONRPCClient(uris[0])

		// Perform import
		return performImport(ctx, hscli, hdcli, hws, ncli, ids.Empty, factory)
	},
}

var exportAssetCmd = &cobra.Command{
	Use: "export-asset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		currentChainID, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to send
		assetID, err := handler.Root().PromptAsset("assetID", true)
		if err != nil {
			return err
		}
		_, decimals, balance, sourceChainID, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetID, true)
		if balance == 0 || err != nil {
			return err
		}

		// Select recipient
		recipient, err := handler.Root().PromptAddress("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := handler.Root().PromptAmount("amount", decimals, balance, nil)
		if err != nil {
			return err
		}

		// Determine return
		var ret bool
		if sourceChainID != ids.Empty {
			ret = true
		}

		// Select reward
		reward, err := handler.Root().PromptAmount("reward", decimals, balance-amount, nil)
		if err != nil {
			return err
		}

		// Determine destination
		destination := sourceChainID
		if !ret {
			destination, _, err = handler.Root().PromptChain("destination", set.Of(currentChainID))
			if err != nil {
				return err
			}
		}

		// Determine if swap in
		swap, err := handler.Root().PromptBool("swap on import")
		if err != nil {
			return err
		}
		var (
			swapIn     uint64
			assetOut   ids.ID
			swapOut    uint64
			swapExpiry int64
		)
		if swap {
			swapIn, err = handler.Root().PromptAmount("swap in", decimals, amount, nil)
			if err != nil {
				return err
			}
			assetOut, err = handler.Root().PromptAsset("asset out (on destination)", true)
			if err != nil {
				return err
			}
			uris, err := handler.Root().GetChain(destination)
			if err != nil {
				return err
			}
			networkID, _, _, err := hcli.Network(ctx)
			if err != nil {
				return err
			}
			dcli := nrpc.NewJSONRPCClient(uris[0], networkID, destination)
			_, decimals, _, _, err := handler.GetAssetInfo(ctx, dcli, priv.Address, assetOut, false)
			if err != nil {
				return err
			}
			swapOut, err = handler.Root().PromptAmount(
				"swap out (on destination, no decimals)",
				decimals,
				hconsts.MaxUint64,
				nil,
			)
			if err != nil {
				return err
			}
			swapExpiry, err = handler.Root().PromptTime("swap expiry")
			if err != nil {
				return err
			}
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		success, txID, err := sendAndWait(ctx, nil, &actions.ExportAsset{
			To:          recipient,
			Asset:       assetID,
			Value:       amount,
			Return:      ret,
			Reward:      reward,
			SwapIn:      swapIn,
			AssetOut:    assetOut,
			SwapOut:     swapOut,
			SwapExpiry:  swapExpiry,
			Destination: destination,
		}, hcli, hws, ncli, factory, true)
		if err != nil {
			return err
		}
		if !success {
			return errors.New("not successful")
		}

		// Perform import
		imp, err := handler.Root().PromptBool("perform import on destination")
		if err != nil {
			return err
		}
		if imp {
			uris, err := handler.Root().GetChain(destination)
			if err != nil {
				return err
			}
			networkID, _, _, err := hcli.Network(ctx)
			if err != nil {
				return err
			}
			hdcli, err := hrpc.NewWebSocketClient(uris[0], hrpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
			if err != nil {
				return err
			}
			if err := performImport(ctx, hcli, hrpc.NewJSONRPCClient(uris[0]), hdcli, nrpc.NewJSONRPCClient(uris[0], networkID, destination), txID, factory); err != nil {
				return err
			}
		}

		// Ask if user would like to switch to destination chain
		sw, err := handler.Root().PromptBool("switch default chain to destination")
		if err != nil {
			return err
		}
		if !sw {
			return nil
		}
		return handler.Root().StoreDefaultChain(destination)
	},
}

// Define the layout that matches the provided date string
// Note: the reference time is "Mon Jan 2 15:04:05 MST 2006" in Go
const (
	TimeLayout = "2006-01-02 15:04:05"
	Auto       = "auto"
	Manual     = "manual"
)

var registerValidatorStakeCmd = &cobra.Command{
	Use: "register-validator-stake [manual | auto <node#>]",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || (args[0] != Manual && args[0] != Auto) {
			return ErrInvalidArgs
		}
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		autoRegister := args[0] == Auto
		nodeNumber := "node1"
		if len(args) == 2 {
			isValid := regexp.MustCompile(`^node([1-9]|10)$`).MatchString(nodeNumber)
			if !isValid {
				return fmt.Errorf("invalid node number")
			}
			nodeNumber = args[1]
		}

		ctx := context.Background()
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		if autoRegister {
			hutils.Outf("{{yellow}}Loading private key for %s{{/}}\n", nodeNumber)
			validatorSignerKey, err := loadPrivateKey("bls", fmt.Sprintf("/tmp/nuklaivm/nodes/%s-bls/signer.key", nodeNumber))
			if err != nil {
				return err
			}
			validatorSignerAddress := codec.MustAddressBech32(nconsts.HRP, validatorSignerKey.Address)
			nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
			if err != nil {
				return err
			}
			_, _, balance, _, err := handler.GetAssetInfo(ctx, nclients[0], validatorSignerKey.Address, ids.Empty, true)
			if err != nil {
				return err
			}
			hutils.Outf("{{yellow}}Balance of validator signer:{{/}} %s\n", hutils.FormatBalance(balance, nconsts.Decimals))
			if balance < uint64(100*math.Pow10(int(nconsts.Decimals))) {
				hutils.Outf("{{yellow}} You need a minimum of 100 NAI to register a validator{{/}}\n")
				return nil
			}
			// Set the default key to the validator signer key
			hutils.Outf("{{yellow}}Loading validator signer key :{{/}} %s\n", validatorSignerAddress)
			if err := handler.h.StoreKey(validatorSignerKey); err != nil && !errors.Is(err, hyperCli.ErrDuplicate) {
				return err
			}
			if err := handler.h.StoreDefaultKey(validatorSignerKey.Address); err != nil {
				return err
			}
			_, priv, factory, hcli, hws, ncli, err = handler.DefaultActor()
			if err != nil {
				return err
			}
		}

		keyType, _ := getKeyType(priv.Address)
		if keyType != blsKey {
			return fmt.Errorf("actor must be a BLS key")
		}
		secretKey, err := bls.PrivateKeyFromBytes(priv.Bytes)
		if err != nil {
			return err
		}
		publicKey := bls.PublicKeyToBytes(bls.PublicFromPrivateKey(secretKey))
		hutils.Outf("{{yellow}}Validator Signer Address: %s\n", codec.MustAddressBech32(nconsts.HRP, priv.Address))

		// Get the validator for which the actor is a signer
		// Get current list of validators
		validators, err := ncli.AllValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		var nodeID ids.NodeID
		for i := 0; i < len(validators); i++ {
			if bytes.Equal(publicKey, validators[i].PublicKey) {
				nodeID = validators[i].NodeID
				break
			}
		}
		hutils.Outf("{{yellow}}Validator NodeID:{{/}} %s\n", nodeID.String())
		if nodeID.Compare(ids.EmptyNodeID) == 0 {
			hutils.Outf("{{red}}actor is not a signer for any of the validators{{/}}\n")
			return nil
		}

		// Get balance info
		_, _, balance, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, ids.Empty, true)
		if balance == 0 || err != nil {
			return err
		}
		if balance < uint64(100*math.Pow10(int(nconsts.Decimals))) {
			hutils.Outf("{{yellow}} You need a minimum of 100 NAI to register a validator{{/}}\n")
			return nil
		}

		// Select staked amount
		stakedAmount, err := handler.Root().PromptAmount("Staked amount", nconsts.Decimals, balance, nil)
		if err != nil {
			return err
		}

		// Get current time
		currentTime := time.Now().UTC()
		stakeStartTime := currentTime.Add(2 * time.Minute)
		stakeEndTime := currentTime.Add(15 * time.Minute)
		delegationFeeRate := 50
		rewardAddress := priv.Address

		if !autoRegister {
			// Select stakeStartTime
			stakeStartTimeString, err := handler.Root().PromptString(
				fmt.Sprintf("Staking Start Time(must be after %s) [YYYY-MM-DD HH:MM:SS]", currentTime.Format(TimeLayout)),
				1,
				32,
			)
			if err != nil {
				return err
			}
			stakeStartTime, err = time.Parse(TimeLayout, stakeStartTimeString)
			if err != nil {
				return err
			}

			// Select stakeEndTime
			stakeEndTimeString, err := handler.Root().PromptString(
				fmt.Sprintf("Staking End Time(must be after %s) [YYYY-MM-DD HH:MM:SS]", stakeStartTimeString),
				1,
				32,
			)
			if err != nil {
				return err
			}
			stakeEndTime, err = time.Parse(TimeLayout, stakeEndTimeString)
			if err != nil {
				return err
			}

			// Select delegationFeeRate
			delegationFeeRate, err = handler.Root().PromptInt("Delegation Fee Rate(must be over 2)", 100)
			if err != nil {
				return err
			}

			// Select rewardAddress
			rewardAddress, err = handler.Root().PromptAddress("Reward Address")
			if err != nil {
				return err
			}
		}

		if stakeStartTime.Before(currentTime) {
			return fmt.Errorf("staking start time must be after the current time (%s)", currentTime.Format(TimeLayout))
		}
		if stakeEndTime.Before(stakeStartTime) {
			return fmt.Errorf("staking end time must be after the staking start time (%s)", stakeStartTime.Format(TimeLayout))
		}
		if delegationFeeRate < 2 || delegationFeeRate > 100 {
			return fmt.Errorf("delegation fee rate must be over 2 and under 100")
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		hutils.Outf("{{yellow}}Register Validator Stake Info - stakeStartTime: %s stakeEndTime: %s delegationFeeRate: %d rewardAddress: %s\n", stakeStartTime.Format(TimeLayout), stakeEndTime.Format(TimeLayout), delegationFeeRate, codec.MustAddressBech32(nconsts.HRP, rewardAddress))

		stakeInfo := &actions.ValidatorStakeInfo{
			NodeID:            nodeID.Bytes(),
			StakeStartTime:    uint64(stakeStartTime.Unix()),
			StakeEndTime:      uint64(stakeEndTime.Unix()),
			StakedAmount:      stakedAmount,
			DelegationFeeRate: uint64(delegationFeeRate),
			RewardAddress:     rewardAddress,
		}
		stakeInfoBytes, err := stakeInfo.Marshal()
		if err != nil {
			return err
		}
		authFactory := auth.NewBLSFactory(secretKey)
		signature, err := authFactory.Sign(stakeInfoBytes)
		if err != nil {
			return err
		}
		signaturePacker := codec.NewWriter(signature.Size(), signature.Size())
		signature.Marshal(signaturePacker)
		authSignature := signaturePacker.Bytes()

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.RegisterValidatorStake{
			StakeInfo:     stakeInfoBytes,
			AuthSignature: authSignature,
		}, hcli, hws, ncli, factory, true)
		return err
	},
}

var getValidatorStakeCmd = &cobra.Command{
	Use: "get-validator-stake",
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			hutils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("validator to get staking info for", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get validator stake
		_, _, _, _, _, _, err = handler.GetValidatorStake(ctx, ncli, nodeID)
		if err != nil {
			return err
		}

		return nil
	},
}

var claimValidatorStakeRewardCmd = &cobra.Command{
	Use: "claim-validator-stake-reward",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			hutils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("validator to claim staking rewards for", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get stake info
		_, _, stakedAmount, _, _, _, err := ncli.ValidatorStake(ctx, nodeID)
		if err != nil {
			return err
		}

		if stakedAmount == 0 {
			hutils.Outf("{{red}}validator has not yet staked{{/}}\n")
			return nil
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.ClaimValidatorStakeRewards{
			NodeID: nodeID.Bytes(),
		}, hcli, hws, ncli, factory, true)
		return err
	},
}

var withdrawValidatorStakeCmd = &cobra.Command{
	Use: "withdraw-validator-stake",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			hutils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("validator to withdraw from staking", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get stake info
		_, _, stakedAmount, _, _, _, err := ncli.ValidatorStake(ctx, nodeID)
		if err != nil {
			return err
		}

		if stakedAmount == 0 {
			hutils.Outf("{{red}}validator has not yet been staked{{/}}\n")
			return nil
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.WithdrawValidatorStake{
			NodeID: nodeID.Bytes(),
		}, hcli, hws, ncli, factory, true)
		return err
	},
}

var delegateUserStakeCmd = &cobra.Command{
	Use: "delegate-user-stake [manual | auto]",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || (args[0] != Manual && args[0] != Auto) {
			return ErrInvalidArgs
		}
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		autoRegister := args[0] == Auto
		ctx := context.Background()
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}
		hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))

		for i := 0; i < len(validators); i++ {
			hutils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("validator to delegate to", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get balance info
		_, _, balance, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, ids.Empty, true)
		if balance == 0 || err != nil {
			return err
		}

		// Select staked amount
		stakedAmount, err := handler.Root().PromptAmount("Staked amount", nconsts.Decimals, balance, nil)
		if err != nil {
			return err
		}

		// Get current time
		currentTime := time.Now().UTC()
		stakeStartTime := currentTime.Add(2 * time.Minute)
		rewardAddress := priv.Address

		if !autoRegister {
			// Select stakeStartTime
			stakeStartTimeString, err := handler.Root().PromptString(
				fmt.Sprintf("Staking Start Time(must be after %s) [YYYY-MM-DD HH:MM:SS]", currentTime.Format(TimeLayout)),
				1,
				32,
			)
			if err != nil {
				return err
			}
			stakeStartTime, err = time.Parse(TimeLayout, stakeStartTimeString)
			if err != nil {
				return err
			}

			// Select rewardAddress
			rewardAddress, err = handler.Root().PromptAddress("Reward Address")
			if err != nil {
				return err
			}
		}

		if stakeStartTime.Before(currentTime) {
			return fmt.Errorf("staking start time must be after the current time (%s)", currentTime.Format(TimeLayout))
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.DelegateUserStake{
			NodeID:         nodeID.Bytes(),
			StakeStartTime: uint64(stakeStartTime.Unix()),
			StakedAmount:   stakedAmount,
			RewardAddress:  rewardAddress,
		}, hcli, hws, ncli, factory, true)
		return err
	},
}

var getUserStakeCmd = &cobra.Command{
	Use: "get-user-stake [address]",
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()

		var address codec.Address
		if len(args) == 0 {
			_, priv, _, _, _, _, err := handler.DefaultActor()
			if err != nil {
				return err
			}
			address = priv.Address
		} else {
			addr, err := codec.ParseAddressBech32(nconsts.HRP, args[0])
			if err != nil {
				return err
			}
			address = addr
		}

		// Get clients
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		ncli := nclients[0]

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			hutils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("validator to get staking info for", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get user stake
		_, _, _, _, err = handler.GetUserStake(ctx, ncli, address, nodeID)
		if err != nil {
			return err
		}

		return nil
	},
}

var claimUserStakeRewardCmd = &cobra.Command{
	Use: "claim-user-stake-reward",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			hutils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("validator to claim staking rewards from", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get stake info
		_, stakedAmount, _, _, err := ncli.UserStake(ctx, priv.Address, nodeID)
		if err != nil {
			return err
		}

		if stakedAmount == 0 {
			hutils.Outf("{{red}}user has not yet delegated to this validator{{/}}\n")
			return nil
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.ClaimDelegationStakeRewards{
			NodeID:           nodeID.Bytes(),
			UserStakeAddress: priv.Address,
		}, hcli, hws, ncli, factory, true)
		return err
	},
}

var undelegateUserStakeCmd = &cobra.Command{
	Use: "undelegate-user-stake",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			hutils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		hutils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			hutils.Outf(
				"{{yellow}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := handler.Root().PromptChoice("validator to unstake from", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get stake info
		_, stakedAmount, _, _, err := ncli.UserStake(ctx, priv.Address, nodeID)
		if err != nil {
			return err
		}

		if stakedAmount == 0 {
			hutils.Outf("{{red}}user has not yet delegated to this validator{{/}}\n")
			return nil
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, _, err = sendAndWait(ctx, nil, &actions.UndelegateUserStake{
			NodeID: nodeID.Bytes(),
		}, hcli, hws, ncli, factory, true)
		return err
	},
}
