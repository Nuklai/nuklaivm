// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"fmt"
	"time"

	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/spf13/cobra"
)

const (
	fsModeWrite     = 0o600
	defaultDatabase = ".nuklai-cli"
	defaultGenesis  = "genesis.json"
)

var (
	handler *Handler

	dbPath                string
	genesisFile           string
	minBlockGap           int64
	minUnitPrice          []string
	maxBlockUnits         []string
	windowTargetUnits     []string
	hideTxs               bool
	randomRecipient       bool
	maxTxBacklog          int
	checkAllChains        bool
	prometheusBaseURI     string
	prometheusOpenBrowser bool
	prometheusFile        string
	prometheusData        string
	startPrometheus       bool
	maxFee                int64
	numCores              int

	rootCmd = &cobra.Command{
		Use:        "nuklai-cli",
		Short:      "NuklaiVM CLI",
		SuggestFor: []string{"nuklai-cli", "nuklaicli"},
	}
)

func init() {
	cobra.EnablePrefixMatching = true
	rootCmd.AddCommand(
		genesisCmd,
		keyCmd,
		chainCmd,
		actionCmd,
		emissionCmd,
		spamCmd,
		prometheusCmd,
	)
	rootCmd.PersistentFlags().StringVar(
		&dbPath,
		"database",
		defaultDatabase,
		"path to database (will create it missing)",
	)
	rootCmd.PersistentPreRunE = func(*cobra.Command, []string) error {
		utils.Outf("{{yellow}}database:{{/}} %s\n", dbPath)
		controller := NewController(dbPath)
		root, err := cli.New(controller)
		if err != nil {
			return err
		}
		handler = NewHandler(root)
		return err
	}
	rootCmd.PersistentPostRunE = func(*cobra.Command, []string) error {
		return handler.Root().CloseDatabase()
	}
	rootCmd.SilenceErrors = true

	// genesis
	genGenesisCmd.PersistentFlags().StringVar(
		&genesisFile,
		"genesis-file",
		defaultGenesis,
		"genesis file path",
	)
	genGenesisCmd.PersistentFlags().StringSliceVar(
		&minUnitPrice,
		"min-unit-price",
		[]string{},
		"minimum price",
	)
	genGenesisCmd.PersistentFlags().StringSliceVar(
		&maxBlockUnits,
		"max-block-units",
		[]string{},
		"max block units",
	)
	genGenesisCmd.PersistentFlags().StringSliceVar(
		&windowTargetUnits,
		"window-target-units",
		[]string{},
		"window target units",
	)
	genGenesisCmd.PersistentFlags().Int64Var(
		&minBlockGap,
		"min-block-gap",
		-1,
		"minimum block gap (ms)",
	)
	genesisCmd.AddCommand(
		genGenesisCmd,
	)

	// key
	balanceKeyCmd.PersistentFlags().BoolVar(
		&checkAllChains,
		"check-all-chains",
		false,
		"check all chains",
	)
	balanceKeyCmd.PersistentFlags().IntVar(
		&numCores,
		"num-cores",
		4,
		"num-cores",
	)
	keyCmd.AddCommand(
		genKeyCmd,
		importKeyCmd,
		setKeyCmd,
		balanceKeyCmd,
		vanityAddressCmd,
	)

	// chain
	watchChainCmd.PersistentFlags().BoolVar(
		&hideTxs,
		"hide-txs",
		false,
		"hide txs",
	)
	chainCmd.AddCommand(
		importChainCmd,
		importANRChainCmd,
		importAvalancheCliChainCmd,
		setChainCmd,
		chainInfoCmd,
		watchChainCmd,
	)

	// actions
	actionCmd.AddCommand(
		transferCmd,

		createAssetCmd,
		mintAssetCmd,
		// burnAssetCmd,
		importAssetCmd,
		exportAssetCmd,

		registerValidatorStakeCmd,
		getValidatorStakeCmd,
		claimValidatorStakeRewardCmd,
		withdrawValidatorStakeCmd,

		delegateUserStakeCmd,
		getUserStakeCmd,
		claimUserStakeRewardCmd,
		undelegateUserStakeCmd,
	)

	emissionModifyCmd.Flags().String("update-emission", "", "Update the tmp/emission-balancer file")
	emissionModifyCmd.Flags().String("address", "", "New emission account address")
	emissionModifyCmd.Flags().Uint64("maxsupply", 0, "New emission max supply")
	emissionModifyCmd.Flags().Uint64("base-apr", 0, "New emission tracker base apr")
	emissionModifyCmd.Flags().Uint64("base-validators", 0, "New emission tracker base validators")
	emissionModifyCmd.Flags().Uint64("epoch-length", 0, "New emission tracker epoch length")

	// emission
	emissionCmd.AddCommand(
		emissionInfoCmd,
		emissionAllValidatorsCmd,
		emissionStakedValidatorsCmd,
		emissionModifyCmd,
	)

	// spam
	runSpamCmd.PersistentFlags().BoolVar(
		&randomRecipient,
		"random-recipient",
		false,
		"random recipient",
	)
	runSpamCmd.PersistentFlags().IntVar(
		&maxTxBacklog,
		"max-tx-backlog",
		72_000,
		"max tx backlog",
	)
	runSpamCmd.PersistentFlags().Int64Var(
		&maxFee,
		"max-fee",
		-1,
		"max fee per tx",
	)
	spamCmd.AddCommand(
		runSpamCmd,
	)

	// prometheus
	generatePrometheusCmd.PersistentFlags().StringVar(
		&prometheusBaseURI,
		"prometheus-base-uri",
		"http://localhost:9090",
		"prometheus server location",
	)
	generatePrometheusCmd.PersistentFlags().BoolVar(
		&prometheusOpenBrowser,
		"prometheus-open-browser",
		true,
		"open browser to prometheus dashboard",
	)
	generatePrometheusCmd.PersistentFlags().StringVar(
		&prometheusFile,
		"prometheus-file",
		"/tmp/prometheus.yaml",
		"prometheus file location",
	)
	generatePrometheusCmd.PersistentFlags().StringVar(
		&prometheusData,
		"prometheus-data",
		fmt.Sprintf("/tmp/prometheus-%d", time.Now().Unix()),
		"prometheus data location",
	)
	generatePrometheusCmd.PersistentFlags().BoolVar(
		&startPrometheus,
		"prometheus-start",
		true,
		"start local prometheus server",
	)
	prometheusCmd.AddCommand(
		generatePrometheusCmd,
	)
}

func Execute() error {
	return rootCmd.Execute()
}
