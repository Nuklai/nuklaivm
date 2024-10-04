// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/near/borsh-go"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/spf13/cobra"
	"github.com/status-im/keycard-go/hexutils"

	"github.com/ava-labs/hypersdk/auth"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/utils"

	hcli "github.com/ava-labs/hypersdk/cli"
	hconsts "github.com/ava-labs/hypersdk/consts"
	nutils "github.com/nuklai/nuklaivm/utils"
)

var actionCmd = &cobra.Command{
	Use: "action",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var transferCmd = &cobra.Command{
	Use: "transfer",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get assetID
		assetID, err := prompt.Asset("assetID", consts.Symbol, true)
		if err != nil {
			return err
		}

		// Get balance info
		balance, _, _, _, decimals, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, assetID, true)
		if balance == 0 || err != nil {
			return err
		}

		// Select recipient
		recipient, err := prompt.Address("recipient")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := parseAmount("amount", decimals, balance)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, txID, err := sendAndWait(ctx, []chain.Action{&actions.Transfer{
			AssetID: assetID.String(),
			To:      recipient,
			Value:   amount,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		utils.Outf("{{yellow}}txID:{{/}} %s\n", txID)
		return processResult(result)
	},
}

var publishFileCmd = &cobra.Command{
	Use: "publishFile",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Select contract bytes
		path, err := prompt.String("contract file", 1, 1000)
		if err != nil {
			return err
		}
		bytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.ContractPublish{
			ContractBytes: bytes,
		}}, cli, bcli, ws, factory)

		if result != nil && result.Success {
			utils.Outf("{{green}}fee consumed:{{/}} %s\n", nutils.FormatBalance(result.Fee, consts.Decimals))

			utils.Outf(hexutils.BytesToHex(result.Outputs[0]) + "\n")
		}
		return err
	},
}

var callCmd = &cobra.Command{
	Use: "call",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get balance info
		balance, err := handler.GetBalance(ctx, bcli, priv.Address, ids.Empty)
		if balance == 0 || err != nil {
			return err
		}

		// Select contract
		contractAddress, err := prompt.Address("contract address")
		if err != nil {
			return err
		}

		// Select amount
		amount, err := parseAmount("amount", consts.Decimals, balance)
		if err != nil {
			return err
		}

		// Select function
		function, err := prompt.String("function", 0, 100)
		if err != nil {
			return err
		}

		action := &actions.ContractCall{
			ContractAddress: contractAddress,
			Value:           amount,
			Function:        function,
		}

		specifiedStateKeysSet, fuel, err := bcli.Simulate(ctx, *action, priv.Address)
		if err != nil {
			return err
		}

		action.SpecifiedStateKeys = make([]actions.StateKeyPermission, 0, len(specifiedStateKeysSet))
		for key, value := range specifiedStateKeysSet {
			action.SpecifiedStateKeys = append(action.SpecifiedStateKeys, actions.StateKeyPermission{Key: key, Permission: value})
		}
		action.Fuel = fuel

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{action}, cli, bcli, ws, factory)
		if result != nil && result.Success {
			utils.Outf("{{green}}fee consumed:{{/}} %s\n", nutils.FormatBalance(result.Fee, consts.Decimals))

			utils.Outf(hexutils.BytesToHex(result.Outputs[0]) + "\n")
			switch function {
			case "balance":
				{
					var intValue uint64
					err := borsh.Deserialize(&intValue, result.Outputs[0])
					if err != nil {
						return err
					}
					utils.Outf("%s\n", nutils.FormatBalance(intValue, consts.Decimals))
				}
			case "get_value":
				{
					var intValue int64
					err := borsh.Deserialize(&intValue, result.Outputs[0])
					if err != nil {
						return err
					}
					utils.Outf("%d\n", intValue)
				}
			}
		}
		return err
	},
}

var deployCmd = &cobra.Command{
	Use: "deploy",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, bcli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		contractID, err := prompt.Bytes("contract id")
		if err != nil {
			return err
		}

		creationInfo, err := prompt.Bytes("creation info")
		if err != nil {
			return err
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.ContractDeploy{
			ContractID:   contractID,
			CreationInfo: creationInfo,
		}}, cli, bcli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
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
		var err error
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		if autoRegister {
			utils.Outf("{{blue}}Loading private key for %s{{/}}\n", nodeNumber)
			validatorSignerKey, err := loadPrivateKeyFromPath("bls", fmt.Sprintf("/tmp/nuklaivm/nodes/%s/signer.key", nodeNumber))
			if err != nil {
				return err
			}
			utils.Outf("{{blue}}Validator Signer Address: %s\n", validatorSignerKey.Address)
			nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
			if err != nil {
				return err
			}
			balance, _, _, _, _, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, nclients[0], validatorSignerKey.Address, ids.Empty, true)
			if err != nil {
				return err
			}
			utils.Outf("{{blue}}Balance of validator signer:{{/}} %s\n", nutils.FormatBalance(balance, consts.Decimals))
			if balance < uint64(100*math.Pow10(int(consts.Decimals))) {
				utils.Outf("{{blue}} You need a minimum of 100 NAI to register a validator{{/}}\n")
				return nil
			}
			// Set the default key to the validator signer key
			utils.Outf("{{blue}}Loading validator signer key :{{/}} %s\n", validatorSignerKey.Address)
			if err := handler.h.StoreKey(validatorSignerKey); err != nil && !errors.Is(err, hcli.ErrDuplicate) {
				return err
			}
			if err := handler.h.StoreDefaultKey(validatorSignerKey.Address); err != nil {
				return err
			}
			_, priv, factory, cli, ncli, ws, err = handler.DefaultActor()
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
		utils.Outf("{{blue}}Validator Signer Address: %s\n", priv.Address)

		// Get the validator for which the actor is a signer
		// Get current list of validators
		validators, err := ncli.AllValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		var nodeID ids.NodeID
		for i := 0; i < len(validators); i++ {
			if bytes.Equal(publicKey, validators[i].PublicKey) {
				nodeID = validators[i].NodeID
				break
			}
		}
		utils.Outf("{{blue}}Validator NodeID:{{/}} %s\n", nodeID.String())
		if nodeID.Compare(ids.EmptyNodeID) == 0 {
			utils.Outf("{{red}}actor is not a signer for any of the validators{{/}}\n")
			return nil
		}

		// Get balance info
		balance, _, _, _, _, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, ids.Empty, true)
		if balance == 0 || err != nil {
			return err
		}
		if balance < uint64(100*math.Pow10(int(consts.Decimals))) {
			utils.Outf("{{blue}} You need a minimum of 100 NAI to register a validator{{/}}\n")
			return nil
		}

		// Select staked amount
		stakedAmount, err := parseAmount("Staked amount", consts.Decimals, balance)
		if err != nil {
			return err
		}

		// Get current block
		currentBlockHeight, _, _, _, _, _, _, err := ncli.EmissionInfo(ctx)
		if err != nil {
			return err
		}

		stakeStartBlock := currentBlockHeight + 15 // roughly 30 seconds from now
		stakeEndBlock := stakeStartBlock + 30*10   // roughly 5 minutes from now
		delegationFeeRate := 50
		rewardAddress := priv.Address

		if !autoRegister {
			// Select stakeStartBlock
			stakeStartBlockInt, err := prompt.Int(
				fmt.Sprintf("Staking Start Block(must be after %d)", currentBlockHeight),
				hconsts.MaxInt,
			)
			if err != nil {
				return err
			}
			stakeStartBlock = uint64(stakeStartBlockInt)

			// Select stakeEndBlock
			stakeEndBlockInt, err := prompt.Int(
				fmt.Sprintf("Staking End Block(must be after %d)", stakeStartBlock),
				hconsts.MaxInt,
			)
			if err != nil {
				return err
			}
			stakeEndBlock = uint64(stakeEndBlockInt)

			// Select delegationFeeRate
			delegationFeeRate, err = prompt.Int("Delegation Fee Rate(must be over 2)", 100)
			if err != nil {
				return err
			}

			// Select rewardAddress
			rewardAddress, err = prompt.Address("Reward Address")
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
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		utils.Outf("{{blue}}Register Validator Stake Info - stakeStartBlock: %d stakeEndBlock: %d delegationFeeRate: %d rewardAddress: %s\n", stakeStartBlock, stakeEndBlock, delegationFeeRate, rewardAddress)

		stakeInfo := &actions.RegisterValidatorStakeResult{
			NodeID:            nodeID,
			StakeStartBlock:   stakeStartBlock,
			StakeEndBlock:     stakeEndBlock,
			StakedAmount:      stakedAmount,
			DelegationFeeRate: uint64(delegationFeeRate),
			RewardAddress:     rewardAddress,
		}
		packer := codec.NewWriter(stakeInfo.Size(), stakeInfo.Size())
		stakeInfo.Marshal(packer)
		stakeInfoBytes := packer.Bytes()
		if packer.Err() != nil {
			return packer.Err()
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
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.RegisterValidatorStake{
			NodeID:        nodeID,
			StakeInfo:     stakeInfoBytes,
			AuthSignature: authSignature,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
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
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{blue}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := prompt.Choice("validator to get staking info for", len(validators))
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
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{blue}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := prompt.Choice("validator to claim staking rewards for", len(validators))
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
			utils.Outf("{{red}}validator has not yet staked{{/}}\n")
			return nil
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.ClaimValidatorStakeRewards{
			NodeID: nodeID,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var withdrawValidatorStakeCmd = &cobra.Command{
	Use: "withdraw-validator-stake",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, _, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{blue}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := prompt.Choice("validator to withdraw from staking", len(validators))
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
			utils.Outf("{{red}}validator has not yet been staked{{/}}\n")
			return nil
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.WithdrawValidatorStake{
			NodeID: nodeID,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
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
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}
		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))

		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{blue}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := prompt.Choice("validator to delegate to", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get balance info
		balance, _, _, _, _, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(ctx, ncli, priv.Address, ids.Empty, true)
		if balance == 0 || err != nil {
			return err
		}

		// Select staked amount
		stakedAmount, err := parseAmount("Staked amount", consts.Decimals, balance)
		if err != nil {
			return err
		}

		// Get current block
		currentBlockHeight, _, _, _, _, _, _, err := ncli.EmissionInfo(ctx)
		if err != nil {
			return err
		}

		stakeStartBlock := currentBlockHeight + 15 // roughly 30 seconds from now
		stakeEndBlock := stakeStartBlock + 30*5    // roughly 5 minutes

		if !autoRegister {
			// Select stakeStartBlock
			stakeStartBlockInt, err := prompt.Int(
				fmt.Sprintf("Staking Start Block(must be after %d)", currentBlockHeight),
				hconsts.MaxInt,
			)
			if err != nil {
				return err
			}
			stakeStartBlock = uint64(stakeStartBlockInt)

			// Select stakeEndBlock
			stakeEndBlockInt, err := prompt.Int(
				fmt.Sprintf("Staking End Block(must be after %d)", stakeStartBlock),
				hconsts.MaxInt,
			)
			if err != nil {
				return err
			}
			stakeEndBlock = uint64(stakeEndBlockInt)
		}

		if stakeStartBlock < currentBlockHeight {
			return fmt.Errorf("staking start block must be after the current block height (%d)", currentBlockHeight)
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.DelegateUserStake{
			NodeID:          nodeID,
			StakeStartBlock: stakeStartBlock,
			StakeEndBlock:   stakeEndBlock,
			StakedAmount:    stakedAmount,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
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
			addr, err := codec.StringToAddress(args[0])
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
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{blue}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := prompt.Choice("validator to get staking info for", len(validators))
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
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{blue}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := prompt.Choice("validator to claim staking rewards from", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get stake info
		_, _, stakedAmount, _, _, err := ncli.UserStake(ctx, priv.Address.String(), nodeID.String())
		if err != nil {
			return err
		}

		if stakedAmount == 0 {
			utils.Outf("{{red}}user has not yet delegated to this validator{{/}}\n")
			return nil
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.ClaimDelegationStakeRewards{
			NodeID: nodeID,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}

var undelegateUserStakeCmd = &cobra.Command{
	Use: "undelegate-user-stake",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()
		_, priv, factory, cli, ncli, ws, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get current list of validators
		validators, err := ncli.StakedValidators(ctx)
		if err != nil {
			return err
		}
		if len(validators) == 0 {
			utils.Outf("{{red}}no validators{{/}}\n")
			return nil
		}

		// Show validators to the user
		utils.Outf("{{cyan}}validators:{{/}} %d\n", len(validators))
		for i := 0; i < len(validators); i++ {
			utils.Outf(
				"{{blue}}%d:{{/}} NodeID=%s\n",
				i,
				validators[i].NodeID,
			)
		}
		// Select validator
		keyIndex, err := prompt.Choice("validator to unstake from", len(validators))
		if err != nil {
			return err
		}
		validatorChosen := validators[keyIndex]
		nodeID := validatorChosen.NodeID

		// Get stake info
		_, _, stakedAmount, _, _, err := ncli.UserStake(ctx, priv.Address.String(), nodeID.String())
		if err != nil {
			return err
		}

		if stakedAmount == 0 {
			utils.Outf("{{red}}user has not yet delegated to this validator{{/}}\n")
			return nil
		}

		// Confirm action
		cont, err := prompt.Continue()
		if !cont || err != nil {
			return err
		}

		// Generate transaction
		result, _, err := sendAndWait(ctx, []chain.Action{&actions.UndelegateUserStake{
			NodeID: nodeID,
		}}, cli, ncli, ws, factory)
		if err != nil {
			return err
		}
		return processResult(result)
	},
}
