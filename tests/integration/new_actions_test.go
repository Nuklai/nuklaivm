// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/ed25519"

	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	nconsts "github.com/nuklai/nuklaivm/consts"
)

func init() {
	logFactory = logging.NewFactory(logging.Config{
		DisplayLevel: logging.Debug,
	})
	l, err := logFactory.Make("main")
	if err != nil {
		panic(err)
	}
	log = l
}

func TestNewActions(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "nuklaivm new actions integration tests")
}

var _ = ginkgo.Describe("nuklai staking mechanism", func() {
	ginkgo.It("Auto register validator stake", func() {
		currentTime := time.Now().UTC()
		stakeStartTime := currentTime.Add(2 * time.Minute)
		stakeEndTime := currentTime.Add(15 * time.Minute)
		delegationFeeRate := 50

		parser0, err := instances[0].ncli.Parser(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())
		parser1, err := instances[1].ncli.Parser(context.Background())
		gomega.Ω(err).Should(gomega.BeNil())

		withdraw0Priv, err := ed25519.GeneratePrivateKey()
		gomega.Ω(err).Should(gomega.BeNil())
		rwithdraw0 := auth.NewED25519Address(withdraw0Priv.PublicKey())
		withdraw0 := codec.MustAddressBech32(nconsts.HRP, rwithdraw0)

		withdraw1Priv, err := ed25519.GeneratePrivateKey()
		gomega.Ω(err).Should(gomega.BeNil())
		// withdraw0Factory := auth.NewED25519Factory(priv1Withdraw)
		withdraw1 := auth.NewED25519Address(withdraw1Priv.PublicKey())

		delegatePriv, err := ed25519.GeneratePrivateKey()
		gomega.Ω(err).Should(gomega.BeNil())
		delegateFactory := auth.NewED25519Factory(delegatePriv)
		rdelegate := auth.NewED25519Address(delegatePriv.PublicKey())
		delegate := codec.MustAddressBech32(nconsts.HRP, rdelegate)

		ginkgo.By("Register validator stake instances[0] with zero balance", func() {
			stakeInfo := &actions.ValidatorStakeInfo{
				NodeID:            instances[0].nodeID.Bytes(),
				StakeStartTime:    uint64(stakeStartTime.Unix()),
				StakeEndTime:      uint64(stakeEndTime.Unix()),
				StakedAmount:      100, // TO DO: SAME TEST WITH 50 TO THROUGH ERROR
				DelegationFeeRate: uint64(delegationFeeRate),
				RewardAddress:     rwithdraw0,
			}
			stakeInfoBytes, err := stakeInfo.Marshal()
			gomega.Ω(err).Should(gomega.BeNil())
			signature, err := factory.Sign(stakeInfoBytes)
			gomega.Ω(err).Should(gomega.BeNil())
			signaturePacker := codec.NewWriter(signature.Size(), signature.Size())
			signature.Marshal(signaturePacker)
			authSignature := signaturePacker.Bytes()
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser0,
				nil,
				&actions.RegisterValidatorStake{
					StakeInfo:     stakeInfoBytes,
					AuthSignature: authSignature,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.HaveOccurred())
			gomega.Ω(submit(context.Background())).ShouldNot(gomega.BeNil())
		})

		ginkgo.By("Get staked validators", func() {
			validators, err := instances[0].ncli.StakedValidators(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(len(validators)).Should(gomega.Equal(1))
		})

		ginkgo.By("Register validator stake instances[1] with zero balance", func() {
			stakeInfo := &actions.ValidatorStakeInfo{
				NodeID:            instances[1].nodeID.Bytes(),
				StakeStartTime:    uint64(stakeStartTime.Unix()),
				StakeEndTime:      uint64(stakeEndTime.Unix()),
				StakedAmount:      100, // TO DO: SAME TEST WITH 50 TO THROUGH ERROR
				DelegationFeeRate: uint64(delegationFeeRate),
				RewardAddress:     withdraw1,
			}
			stakeInfoBytes, err := stakeInfo.Marshal()
			gomega.Ω(err).Should(gomega.BeNil())
			signature, err := factory.Sign(stakeInfoBytes)
			gomega.Ω(err).Should(gomega.BeNil())
			signaturePacker := codec.NewWriter(signature.Size(), signature.Size())
			signature.Marshal(signaturePacker)
			authSignature := signaturePacker.Bytes()
			submit, _, _, err := instances[1].hcli.GenerateTransaction(
				context.Background(),
				parser1,
				nil,
				&actions.RegisterValidatorStake{
					StakeInfo:     stakeInfoBytes,
					AuthSignature: authSignature,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.HaveOccurred())
			gomega.Ω(submit(context.Background())).ShouldNot(gomega.BeNil())
		})

		ginkgo.By("Get validator staked amount after staking", func() {
			_, _, stakedAmount, _, _, _, err := instances[0].ncli.ValidatorStake(context.Background(), instances[0].nodeID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(stakedAmount).Should(gomega.Equal(100))
			_, _, stakedAmount, _, _, _, err = instances[1].ncli.ValidatorStake(context.Background(), instances[1].nodeID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(stakedAmount).Should(gomega.Equal(100))
		})

		ginkgo.By("Get staked validators", func() {
			validators, err := instances[0].ncli.StakedValidators(context.Background())
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(len(validators)).Should(gomega.Equal(2))
		})

		ginkgo.By("Transfer NAI to user and delegate stake to instances[0]", func() {
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser0,
				nil,
				&actions.Transfer{
					To:    rdelegate,
					Asset: ids.Empty,
					Value: 100,
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
			currentTime := time.Now().UTC()
			userStakeStartTime := currentTime.Add(2 * time.Minute)
			gomega.Ω(err).Should(gomega.BeNil())
			submit, _, _, err = instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser0,
				nil,
				&actions.DelegateUserStake{
					NodeID:         instances[0].nodeID.Bytes(),
					StakeStartTime: uint64(userStakeStartTime.Unix()),
					StakedAmount:   50,
					RewardAddress:  rdelegate,
				},
				delegateFactory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		})

		ginkgo.By("Get user stake before claim", func() {
			_, stakedAmount, _, _, err := instances[0].ncli.UserStake(context.Background(), rdelegate, instances[0].nodeID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(stakedAmount).Should(gomega.BeNumerically("==", 50))
		})

		ginkgo.By("Claim delegation stake rewards from instances[0]", func() {
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser0,
				nil,
				&actions.ClaimDelegationStakeRewards{
					NodeID:           instances[0].nodeID.Bytes(),
					UserStakeAddress: rdelegate,
				},
				delegateFactory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		})

		ginkgo.By("Get user stake after claim", func() {
			_, stakedAmount, _, _, err := instances[0].ncli.UserStake(context.Background(), rdelegate, instances[0].nodeID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(stakedAmount).Should(gomega.BeNumerically("==", 0))
		})

		ginkgo.By("Undelegate user stake from instances[0]", func() {
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser0,
				nil,
				&actions.UndelegateUserStake{
					NodeID: instances[0].nodeID.Bytes(),
				},
				delegateFactory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		})

		// add more ginko.By where error should be thrown with wrong data input
		ginkgo.By("Claim validator instances[0] stake reward", func() {
			// ClaimValidatorStakeRewards
			// TO DO: test claim with a wrong key
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser0,
				nil,
				&actions.ClaimValidatorStakeRewards{
					NodeID: instances[0].nodeID.Bytes(),
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(instances[0].ncli.Balance(context.Background(), withdraw0, ids.Empty)).Should(gomega.BeNumerically(">", 0))
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		})

		ginkgo.By("Withdraw validator instances[0] stake", func() {
			// WithdrawValidatorStake
			// TO DO: test claim with a wrong key
			submit, _, _, err := instances[0].hcli.GenerateTransaction(
				context.Background(),
				parser0,
				nil,
				&actions.WithdrawValidatorStake{
					NodeID: instances[0].nodeID.Bytes(),
				},
				factory,
			)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(submit(context.Background())).Should(gomega.BeNil())
		})

		ginkgo.By("Get validator stake after staking withdraw ", func() {
			_, _, stakedAmount, _, _, _, err := instances[0].ncli.ValidatorStake(context.Background(), instances[0].nodeID)
			gomega.Ω(err).Should(gomega.BeNil())
			gomega.Ω(stakedAmount).Should(gomega.Equal(0))
		})
	})
})
