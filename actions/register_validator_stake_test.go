// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package actions

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/auth"
	"github.com/ava-labs/hypersdk/chain/chaintest"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/state"
)

func TestRegisterValidatorStakeAction(t *testing.T) {
	nodeID := ids.GenerateTestNodeID()
	otherNodeID := ids.GenerateTestNodeID()

	// Mock valid stake information
	stakeInfo1, authSignature1, privateKey, publicKey := generateStakeInfoAndSignature(nodeID, 60, 200, emission.GetStakingConfig().MinValidatorStake, 10)
	stakeInfo2, authSignature2, _, _ := generateStakeInfoAndSignature(nodeID, 60, 200, 10, 10)

	addr := privateKey.Address
	emission.MockNewEmission(&emission.MockEmission{
		LastAcceptedBlockHeight: 50, // Mock block height
		Validator: &emission.Validator{
			IsActive:          true,
			NodeID:            nodeID,
			PublicKey:         publicKey,
			StakedAmount:      emission.GetStakingConfig().MinValidatorStake,
			DelegationFeeRate: 10,
		},
	})

	tests := []chaintest.ActionTest{
		{
			Name:  "InvalidNodeID",
			Actor: addr,
			Action: &RegisterValidatorStake{
				NodeID:        otherNodeID, // Different NodeID than the one in StakeInfo
				StakeInfo:     stakeInfo1,
				AuthSignature: authSignature1,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrInvalidNodeID,
		},
		{
			Name:  "ValidatorAlreadyRegistered",
			Actor: addr,
			Action: &RegisterValidatorStake{
				NodeID:        nodeID,
				StakeInfo:     stakeInfo1,
				AuthSignature: authSignature1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Register the validator
				require.NoError(t, storage.SetRegisterValidatorStake(context.Background(), store, nodeID, 50, 150, 10000, 10, addr, addr))
				return store
			}(),
			ExpectedErr: ErrValidatorAlreadyRegistered,
		},
		{
			Name:  "InvalidStakeAmount",
			Actor: addr,
			Action: &RegisterValidatorStake{
				NodeID:        nodeID,
				StakeInfo:     stakeInfo2,
				AuthSignature: authSignature2,
			},
			State:       chaintest.NewInMemoryStore(),
			ExpectedErr: ErrValidatorStakedAmountInvalid,
		},
		{
			Name:     "ValidRegisterValidatorStake",
			ActionID: ids.GenerateTestID(),
			Actor:    addr,
			Action: &RegisterValidatorStake{
				NodeID:        nodeID,
				StakeInfo:     stakeInfo1,
				AuthSignature: authSignature1,
			},
			State: func() state.Mutable {
				store := chaintest.NewInMemoryStore()
				// Set the balance for the user
				require.NoError(t, storage.SetBalance(context.Background(), store, addr, ids.Empty, emission.GetStakingConfig().MinValidatorStake*2))
				return store
			}(),
			Assertion: func(ctx context.Context, t *testing.T, store state.Mutable) {
				// Check if balance is correctly deducted
				balance, err := storage.GetBalance(ctx, store, addr, ids.Empty)
				require.NoError(t, err)
				require.Equal(t, emission.GetStakingConfig().MinValidatorStake, balance)

				// Check if the stake was created correctly
				exists, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate, rewardAddress, ownerAddress, _ := storage.GetRegisterValidatorStake(ctx, store, nodeID)
				require.True(t, exists)
				require.Equal(t, uint64(60), stakeStartBlock)
				require.Equal(t, uint64(200), stakeEndBlock)
				require.Equal(t, emission.GetStakingConfig().MinValidatorStake, stakedAmount)
				require.Equal(t, uint64(10), delegationFeeRate)
				require.Equal(t, addr, rewardAddress)
				require.Equal(t, addr, ownerAddress)
			},
			ExpectedOutputs: &RegisterValidatorStakeResult{
				NodeID:            nodeID,
				StakeStartBlock:   60,
				StakeEndBlock:     200,
				StakedAmount:      emission.GetStakingConfig().MinValidatorStake,
				DelegationFeeRate: 10,
				RewardAddress:     addr,
			},
		},
	}

	for _, tt := range tests {
		tt.Run(context.Background(), t)
	}
}

func BenchmarkRegisterValidatorStake(b *testing.B) {
	require := require.New(b)

	nodeID := ids.GenerateTestNodeID()

	// Mock valid stake information
	stakeInfo1, authSignature1, privateKey, publicKey := generateStakeInfoAndSignature(nodeID, 60, 200, emission.GetStakingConfig().MinValidatorStake, 10)

	actor := privateKey.Address
	emission.MockNewEmission(&emission.MockEmission{
		LastAcceptedBlockHeight: 50, // Mock block height
		Validator: &emission.Validator{
			IsActive:          true,
			NodeID:            nodeID,
			PublicKey:         publicKey,
			StakedAmount:      emission.GetStakingConfig().MinValidatorStake,
			DelegationFeeRate: 10,
		},
	})

	registerValidatorStakeBenchmark := &chaintest.ActionBenchmark{
		Name:  "RegisterValidatorStakeBenchmark",
		Actor: actor,
		Action: &RegisterValidatorStake{
			NodeID:        nodeID,
			StakeInfo:     stakeInfo1,
			AuthSignature: authSignature1,
		},
		CreateState: func() state.Mutable {
			store := chaintest.NewInMemoryStore()
			// Set the balance for the user
			require.NoError(storage.SetBalance(context.Background(), store, actor, ids.Empty, emission.GetStakingConfig().MinValidatorStake*2))
			return store
		},
		Assertion: func(ctx context.Context, b *testing.B, store state.Mutable) {
			// Check if balance is correctly deducted
			balance, err := storage.GetBalance(ctx, store, actor, ids.Empty)
			require.NoError(err)
			require.Equal(b, emission.GetStakingConfig().MinValidatorStake, balance)

			// Check if the stake was created correctly
			exists, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate, rewardAddress, ownerAddress, _ := storage.GetRegisterValidatorStake(ctx, store, nodeID)
			require.True(exists)
			require.Equal(b, uint64(60), stakeStartBlock)
			require.Equal(b, uint64(200), stakeEndBlock)
			require.Equal(b, emission.GetStakingConfig().MinValidatorStake, stakedAmount)
			require.Equal(b, uint64(10), delegationFeeRate)
			require.Equal(b, actor, rewardAddress)
			require.Equal(b, actor, ownerAddress)
		},
	}

	ctx := context.Background()
	registerValidatorStakeBenchmark.Run(ctx, b)
}

func generateStakeInfoAndSignature(nodeID ids.NodeID, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate uint64) ([]byte, []byte, *cli.PrivateKey, []byte) {
	blsBase64Key, err := base64.StdEncoding.DecodeString("MdWjv5OOW/p/JKt673vYxwROsfTCO7iZ2jwWCnY18hw=")
	if err != nil {
		panic(err)
	}
	secretKey, err := bls.PrivateKeyFromBytes(blsBase64Key)
	if err != nil {
		panic(err)
	}
	blsPrivateKey := &cli.PrivateKey{
		Address: auth.NewBLSAddress(bls.PublicFromPrivateKey(secretKey)),
		Bytes:   bls.PrivateKeyToBytes(secretKey),
	}
	blsPublicKey := bls.PublicKeyToBytes(bls.PublicFromPrivateKey(secretKey))

	stakeInfo := &RegisterValidatorStakeResult{
		NodeID:            nodeID,
		StakeStartBlock:   stakeStartBlock,
		StakeEndBlock:     stakeEndBlock,
		StakedAmount:      stakedAmount,
		DelegationFeeRate: delegationFeeRate,
		RewardAddress:     blsPrivateKey.Address,
	}
	packer := codec.NewWriter(stakeInfo.Size(), stakeInfo.Size())
	stakeInfo.Marshal(packer)
	stakeInfoBytes := packer.Bytes()
	if err != nil {
		panic(packer.Err())
	}
	authFactory := auth.NewBLSFactory(secretKey)
	signature, err := authFactory.Sign(stakeInfoBytes)
	if err != nil {
		panic(err)
	}
	signaturePacker := codec.NewWriter(signature.Size(), signature.Size())
	signature.Marshal(signaturePacker)
	authSignature := signaturePacker.Bytes()
	return stakeInfoBytes, authSignature, blsPrivateKey, blsPublicKey
}
