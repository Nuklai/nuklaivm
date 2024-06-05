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
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	hyperCli "github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/bls"
	hutils "github.com/ava-labs/hypersdk/utils"

	frpc "github.com/nuklai/nuklai-faucet/rpc"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	nconsts "github.com/nuklai/nuklaivm/consts"
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
		if _, err = sendAndWait(ctx, []chain.Action{&actions.Transfer{
			To:    addr,
			Asset: ids.Empty,
			Value: amount,
		}}, cli, scli, tcli, factory, true); err != nil {
			return err
		}
		hutils.Outf("{{green}}funded faucet:{{/}} %s\n", faucetAddress)
		return nil
	},
}

var transferCmd = &cobra.Command{
	Use: "transfer",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			assetID   ids.ID
			recipient codec.Address
			amount    uint64
			err       error
		)

		ctx := context.Background()
		_, priv, factory, hcli, hws, ncli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get assetID
		assetIDStr, _ := cmd.Flags().GetString("assetID")
		if assetIDStr == "" {
			assetID, err = handler.Root().PromptAsset("assetID", true)
			if err != nil {
				return err
			}
		} else {
			if assetIDStr == nconsts.Symbol {
				assetID = ids.Empty
			} else {
				assetID, err = ids.FromString(assetIDStr)
				if err != nil {
					return err
				}
			}
		}

		_, decimals, balance, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetID, true)
		if balance == 0 || err != nil {
			return err
		}

		// Get recipient
		recipientStr, _ := cmd.Flags().GetString("recipient")
		if recipientStr == "" {
			recipient, err = handler.Root().PromptAddress("recipient")
			if err != nil {
				return err
			}
		} else {
			recipient, err = codec.ParseAddressBech32(nconsts.HRP, recipientStr)
			if err != nil {
				return err
			}
		}

		// Get amount
		amountStr, _ := cmd.Flags().GetString("amount")
		if amountStr == "" {
			amount, err = handler.Root().PromptAmount("amount", decimals, balance, nil)
			if err != nil {
				return err
			}
		} else {
			amount, err = hutils.ParseBalance(amountStr, nconsts.Decimals)
			if err != nil {
				return err
			}
		}

		// Confirm action
		if assetIDStr == "" || recipientStr == "" || amountStr == "" {
			confirm, err := handler.Root().PromptContinue()
			if !confirm || err != nil {
				return errors.New("transfer not confirmed")
			}
		} else {
			// Auto-confirm if all arguments are provided
			hutils.Outf("All arguments provided, auto-confirming the transfer\n")
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.Transfer{
			To:    recipient,
			Asset: assetID,
			Value: amount,
		}}, hcli, hws, ncli, factory, true)
		return err
	},
}

var createAssetCmd = &cobra.Command{
	Use: "create-asset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, scli, tcli, err := handler.DefaultActor()
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
		txID, err := sendAndWait(ctx, []chain.Action{&actions.CreateAsset{
			Symbol:   []byte(symbol),
			Decimals: uint8(decimals), // already constrain above to prevent overflow
			Metadata: []byte(metadata),
		}}, cli, scli, tcli, factory, true)
		if err != nil {
			return err
		}

		// Print assetID
		assetID := chain.CreateActionID(txID, 0)
		hutils.Outf("{{green}}assetID:{{/}} %s\n", assetID)
		return nil
	},
}

var mintAssetCmd = &cobra.Command{
	Use: "mint-asset",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, scli, tcli, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select token to mint
		assetID, err := handler.Root().PromptAsset("assetID", false)
		if err != nil {
			return err
		}
		exists, symbol, decimals, metadata, supply, owner, err := tcli.Asset(ctx, assetID, false)
		if err != nil {
			return err
		}
		if !exists {
			hutils.Outf("{{red}}%s does not exist{{/}}\n", assetID)
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
		amount, err := handler.Root().PromptAmount("amount", decimals, consts.MaxUint64-supply, nil)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.MintAsset{
			Asset: assetID,
			To:    recipient,
			Value: amount,
		}}, cli, scli, tcli, factory, true)
		return err
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

		// Get current block
		currentBlockHeight, _, _, _, _, _, _, err := ncli.EmissionInfo(ctx)
		if err != nil {
			return err
		}

		stakeStartBlock := currentBlockHeight + 20 // roughly 1 minute from now
		stakeEndBlock := stakeStartBlock + 20*2    // roughly 10 minutes from stakeStartBlock
		delegationFeeRate := 50
		rewardAddress := priv.Address

		if !autoRegister {
			// Select stakeStartBlock
			stakeStartBlockString, err := handler.Root().PromptString(
				fmt.Sprintf("Staking Start Block(must be after %d)", currentBlockHeight),
				1,
				32,
			)
			if err != nil {
				return err
			}

			stakeStartBlock, err = strconv.ParseUint(stakeStartBlockString, 10, 64)
			if err != nil {
				return err
			}

			// Select stakeEndBlock
			stakeEndBlockString, err := handler.Root().PromptString(
				fmt.Sprintf("Staking End Block(must be after %s)", stakeStartBlockString),
				1,
				32,
			)
			if err != nil {
				return err
			}
			stakeEndBlock, err = strconv.ParseUint(stakeEndBlockString, 10, 64)
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

		if stakeStartBlock < currentBlockHeight {
			return fmt.Errorf("staking start block must be after the current block height (%d)", currentBlockHeight)
		}
		if stakeEndBlock < stakeStartBlock {
			return fmt.Errorf("staking end block must be after the staking start block height (%d)", stakeStartBlock)
		}
		if delegationFeeRate < 2 || delegationFeeRate > 100 {
			return fmt.Errorf("delegation fee rate must be over 2 and under 100")
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		hutils.Outf("{{yellow}}Register Validator Stake Info - stakeStartBlock: %d stakeEndBlock: %d delegationFeeRate: %d rewardAddress: %s\n", stakeStartBlock, stakeEndBlock, delegationFeeRate, codec.MustAddressBech32(nconsts.HRP, rewardAddress))

		stakeInfo := &actions.ValidatorStakeInfo{
			NodeID:            nodeID.Bytes(),
			StakeStartBlock:   stakeStartBlock,
			StakeEndBlock:     stakeEndBlock,
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
		_, err = sendAndWait(ctx, []chain.Action{&actions.RegisterValidatorStake{
			StakeInfo:     stakeInfoBytes,
			AuthSignature: authSignature,
		}}, hcli, hws, ncli, factory, true)
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
		_, err = sendAndWait(ctx, []chain.Action{&actions.ClaimValidatorStakeRewards{
			NodeID: nodeID.Bytes(),
		}}, hcli, hws, ncli, factory, true)
		return err
	},
}

var withdrawValidatorStakeCmd = &cobra.Command{
	Use: "withdraw-validator-stake",
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
		_, err = sendAndWait(ctx, []chain.Action{&actions.WithdrawValidatorStake{
			NodeID:        nodeID.Bytes(),
			RewardAddress: priv.Address,
		}}, hcli, hws, ncli, factory, true)
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

		// Get current block
		currentBlockHeight, _, _, _, _, _, _, err := ncli.EmissionInfo(ctx)
		if err != nil {
			return err
		}

		stakeStartBlock := currentBlockHeight + 20 // roughly 1 minute from now
		stakeEndBlock := stakeStartBlock + 20*2    // roughly 10 minutes from stakeStartBlock
		rewardAddress := priv.Address

		if !autoRegister {
			// Select stakeStartBlock
			stakeStartBlockString, err := handler.Root().PromptString(
				fmt.Sprintf("Staking Start Block(must be after %d)", currentBlockHeight),
				1,
				32,
			)
			if err != nil {
				return err
			}

			stakeStartBlock, err = strconv.ParseUint(stakeStartBlockString, 10, 64)
			if err != nil {
				return err
			}

			// Select stakeEndBlock
			stakeEndBlockString, err := handler.Root().PromptString(
				fmt.Sprintf("Staking End Block(must be after %s)", stakeStartBlockString),
				1,
				32,
			)
			if err != nil {
				return err
			}
			stakeEndBlock, err = strconv.ParseUint(stakeEndBlockString, 10, 64)
			if err != nil {
				return err
			}

			// Select rewardAddress
			rewardAddress, err = handler.Root().PromptAddress("Reward Address")
			if err != nil {
				return err
			}
		}

		if stakeStartBlock < currentBlockHeight {
			return fmt.Errorf("staking start block must be after the current block height (%d)", currentBlockHeight)
		}

		// Confirm action
		cont, err := handler.Root().PromptContinue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		_, err = sendAndWait(ctx, []chain.Action{&actions.DelegateUserStake{
			NodeID:          nodeID.Bytes(),
			StakeStartBlock: stakeStartBlock,
			StakeEndBlock:   stakeEndBlock,
			StakedAmount:    stakedAmount,
			RewardAddress:   rewardAddress,
		}}, hcli, hws, ncli, factory, true)
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
		_, _, _, _, _, err = handler.GetUserStake(ctx, ncli, address, nodeID)
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
		_, _, stakedAmount, _, _, err := ncli.UserStake(ctx, priv.Address, nodeID)
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
		_, err = sendAndWait(ctx, []chain.Action{&actions.ClaimDelegationStakeRewards{
			NodeID:           nodeID.Bytes(),
			UserStakeAddress: priv.Address,
		}}, hcli, hws, ncli, factory, true)
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
		_, _, stakedAmount, _, _, err := ncli.UserStake(ctx, priv.Address, nodeID)
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
		_, err = sendAndWait(ctx, []chain.Action{&actions.UndelegateUserStake{
			NodeID:        nodeID.Bytes(),
			RewardAddress: priv.Address,
		}}, hcli, hws, ncli, factory, true)
		return err
	},
}
