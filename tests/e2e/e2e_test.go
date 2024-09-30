// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package e2e_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ava-labs/avalanchego/tests"
	"github.com/ava-labs/avalanchego/tests/fixture/e2e"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/tests/workload"
	"github.com/nuklai/nuklaivm/vm"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/hypersdk/abi"
	"github.com/ava-labs/hypersdk/tests/fixture"

	he2e "github.com/ava-labs/hypersdk/tests/e2e"
	ginkgo "github.com/onsi/ginkgo/v2"
)

const owner = "nuklaivm-e2e-tests"

var flagVars *e2e.FlagVars

func TestE2e(t *testing.T) {
	ginkgo.RunSpecs(t, "nuklaivm e2e test suites")
}

func init() {
	flagVars = e2e.RegisterFlags()
}

// Construct tmpnet network with a single VMWithContracts Subnet
var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	require := require.New(ginkgo.GinkgoT())

	gen, workloadFactory, err := workload.New(100 /* minBlockGap: 100ms */)
	require.NoError(err)

	genesisBytes, err := json.Marshal(gen)
	require.NoError(err)

	expectedABI, err := abi.NewABI(vm.ActionParser.GetRegisteredTypes(), vm.OutputParser.GetRegisteredTypes())
	require.NoError(err)

	// Import HyperSDK e2e test coverage and inject VMWithContracts name
	// and workload factory to orchestrate the test.
	he2e.SetWorkload(consts.Name, workloadFactory, expectedABI)

	tc := e2e.NewTestContext()

	testEnv := fixture.NewTestEnvironment(tc, flagVars, owner, consts.Name, consts.ID, genesisBytes).Marshal()

	// Convert testEnv to TestEnvironment
	var env e2e.TestEnvironment
	err = json.Unmarshal(testEnv, &env)
	require.NoError(err)

	// Print the network configuration to the console
	tc.Outf("Network Dir: %s\n", env.NetworkDir)

	// Copy the signer key content to the destination directory
	err = copyStakingSignerKey(tc, env.NetworkDir)
	require.NoError(err)

	return testEnv
}, func(envBytes []byte) {
	// Run in every ginkgo process

	// Initialize the local test environment from the global state
	e2e.InitSharedTestEnvironment(ginkgo.GinkgoT(), envBytes)
})

type Flags struct {
	StakingSignerKeyFileContent string `json:"staking-signer-key-file-content"`
}

// Function to copy the signer key content from the flags.json file and create the signer.json file
func copyStakingSignerKey(tc tests.TestContext, networkDir string) error {
	// List all directories in networkDir that start with "NodeID-"
	directories, err := os.ReadDir(networkDir)
	if err != nil {
		return fmt.Errorf("failed to read network directory: %w", err)
	}

	for index, dir := range directories {
		if dir.IsDir() && strings.HasPrefix(dir.Name(), "NodeID-") {
			// Construct the path to flags.json within the current NodeID directory
			flagsFilePath := filepath.Join(networkDir, dir.Name(), "flags.json")

			// Read and parse the flags.json file
			flagsData, err := os.ReadFile(flagsFilePath)
			if err != nil {
				return fmt.Errorf("failed to read flags.json file at %s: %w", flagsFilePath, err)
			}

			var flags Flags
			if err := json.Unmarshal(flagsData, &flags); err != nil {
				return fmt.Errorf("failed to unmarshal flags.json at %s: %w", flagsFilePath, err)
			}

			// Decode the staking-signer-key-file-content
			signerKeyBytes, err := base64.StdEncoding.DecodeString(flags.StakingSignerKeyFileContent)
			if err != nil {
				return fmt.Errorf("failed to decode signer key content: %w", err)
			}

			// Define the destination directory and paths using node1, node2, etc.
			destNodeDir := fmt.Sprintf("node%d", index+1)
			destSignerKeyPath := filepath.Join("/tmp/nuklaivm/nodes", destNodeDir, "signer.key")
			destSignerJSONPath := filepath.Join("/tmp/nuklaivm/nodes", destNodeDir, "signer.json")

			// Ensure the destination directory exists, create it if it doesn't
			destDir := filepath.Dir(destSignerKeyPath)
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
			}

			// Write the decoded signer key content to the destination signer.key file
			if err := os.WriteFile(destSignerKeyPath, signerKeyBytes, 0o600); err != nil {
				return fmt.Errorf("failed to write signer key to %s: %w", destSignerKeyPath, err)
			}

			// Create the signer.json file with the base64 and hex representations
			signerJSONContent := map[string]string{
				"privateKeyBase64": flags.StakingSignerKeyFileContent,
				"nodeID":           dir.Name(),
			}

			signerJSONData, err := json.MarshalIndent(signerJSONContent, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal signer.json content: %w", err)
			}

			if err := os.WriteFile(destSignerJSONPath, signerJSONData, 0o600); err != nil {
				return fmt.Errorf("failed to write signer.json to %s: %w", destSignerJSONPath, err)
			}

			tc.Outf("Successfully copied signer key to %s and created signer.json at %s\n", destSignerKeyPath, destSignerJSONPath)
		}
	}
	return nil
}
