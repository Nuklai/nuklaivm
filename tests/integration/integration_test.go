// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package integration_test

import (
	"encoding/json"
	"testing"

	"github.com/nuklai/nuklaivm/vm"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/auth"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/tests/integration"

	lconsts "github.com/nuklai/nuklaivm/consts"
	nuklaivmWorkload "github.com/nuklai/nuklaivm/tests/workload"
	ginkgo "github.com/onsi/ginkgo/v2"
)

func TestIntegration(t *testing.T) {
	ginkgo.RunSpecs(t, "nuklaivm integration test suites")
}

var _ = ginkgo.BeforeSuite(func() {
	require := require.New(ginkgo.GinkgoT())
	genesis, workloadFactory, err := nuklaivmWorkload.New(0, "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9", "00c4cb545f748a28770042f893784ce85b107389004d6a0e0d6d7518eeae1292d9")
	require.NoError(err)

	genesisBytes, err := json.Marshal(genesis)
	require.NoError(err)

	randomEd25519Priv, err := ed25519.GeneratePrivateKey()
	require.NoError(err)

	randomEd25519AuthFactory := auth.NewED25519Factory(randomEd25519Priv)

	// Setup imports the integration test coverage
	integration.Setup(
		vm.New,
		genesisBytes,
		lconsts.ID,
		vm.CreateParser,
		workloadFactory,
		randomEd25519AuthFactory,
	)
})
