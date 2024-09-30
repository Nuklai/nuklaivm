// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/utils"
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
	minUnitPrice          []string
	maxBlockUnits         []string
	windowTargetUnits     []string
	minBlockGap           int64
	hideTxs               bool
	checkAllChains        bool
	prometheusBaseURI     string
	prometheusOpenBrowser bool
	prometheusFile        string
	prometheusData        string
	startPrometheus       bool

	rootCmd = &cobra.Command{
		Use:        "nuklai-cli",
		Short:      "Nuklai CLI",
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
		assetCmd,
		emissionCmd,
		datasetCmd,
		marketplaceCmd,
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
	keyCmd.AddCommand(
		genKeyCmd,
		importKeyCmd,
		setKeyCmd,
		balanceKeyCmd,
		balanceFTKeyCmd,
		balanceNFTKeyCmd,
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
		importCliChainCmd,
		setChainCmd,
		chainInfoCmd,
		watchChainCmd,
	)

	// actions
	actionCmd.AddCommand(
		transferCmd,

		callCmd,
		publishFileCmd,
		deployCmd,

		registerValidatorStakeCmd,
		getValidatorStakeCmd,
		claimValidatorStakeRewardCmd,
		withdrawValidatorStakeCmd,

		delegateUserStakeCmd,
		getUserStakeCmd,
		claimUserStakeRewardCmd,
		undelegateUserStakeCmd,
	)

	// emission
	emissionCmd.AddCommand(
		emissionInfoCmd,
		emissionAllValidatorsCmd,
		emissionStakedValidatorsCmd,
	)

	// asset
	assetCmd.AddCommand(
		createAssetCmd,
		updateAssetCmd,
		mintAssetFTCmd,
		mintAssetNFTCmd,
		burnAssetFTCmd,
		burnAssetNFTCmd,
	)

	// dataset
	datasetCmd.AddCommand(
		createDatasetCmd,
		createDatasetFromExistingAssetCmd,
		updateDatasetCmd,
		getDatasetCmd,
		initiateContributeDatasetCmd,
		getDataContributionPendingCmd,
		completeContributeDatasetCmd,
	)

	// marketplace
	marketplaceCmd.AddCommand(
		publishDatasetMarketplaceCmd,
		subscribeDatasetMarketplaceCmd,
		infoDatasetMarketplaceCmd,
		claimPaymentMarketplaceCmd,
	)

	// spam
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
