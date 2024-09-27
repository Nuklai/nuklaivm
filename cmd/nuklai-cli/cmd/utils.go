// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/vm"
	"gopkg.in/yaml.v2"
)

type ClusterInfo struct {
	ChainID  string `yaml:"CHAIN_ID"` // ids.ID requires "first and last characters to be quotes"
	SubnetID string `yaml:"SUBNET_ID"`
	APIs     []struct {
		CloudID string `yaml:"CLOUD_ID"`
		IP      string `yaml:"IP"`
		Region  string `yaml:"REGION"`
	} `yaml:"API"`
	Validators []struct {
		CloudID string `yaml:"CLOUD_ID"`
		IP      string `yaml:"IP"`
		Region  string `yaml:"REGION"`
		NodeID  string `yaml:"NODE_ID"`
	} `yaml:"VALIDATOR"`
}

func ReadCLIFile(cliPath string) (ids.ID, map[string]string, error) {
	// Load yaml file
	yamlFile, err := os.ReadFile(cliPath)
	if err != nil {
		return ids.Empty, nil, err
	}
	var yamlContents ClusterInfo
	if err := yaml.Unmarshal(yamlFile, &yamlContents); err != nil {
		return ids.Empty, nil, fmt.Errorf("%w: unable to unmarshal YAML", err)
	}
	chainID, err := ids.FromString(yamlContents.ChainID)
	if err != nil {
		return ids.Empty, nil, err
	}

	// Load nodes
	nodes := make(map[string]string)
	for i, api := range yamlContents.APIs {
		name := fmt.Sprintf("%s-%d (%s)", "API", i, api.Region)
		uri := fmt.Sprintf("http://%s:9650/ext/bc/%s", api.IP, chainID)
		nodes[name] = uri
	}
	for i, validator := range yamlContents.Validators {
		name := fmt.Sprintf("%s-%d (%s)", "Validator", i, validator.Region)
		uri := fmt.Sprintf("http://%s:9650/ext/bc/%s", validator.IP, chainID)
		nodes[name] = uri
	}
	return chainID, nodes, nil
}

// Helper function to process the transaction result
func processResult(result *chain.Result) error {
	if result != nil && result.Success {
		utils.Outf("{{green}}fee consumed:{{/}} %s NAI\n", utils.FormatBalance(result.Fee, consts.Decimals))

		// Use NewReader to create a Packer from the result output
		packer := codec.NewReader(result.Outputs[0], len(result.Outputs[0]))
		r, err := vm.OutputParser.Unmarshal(packer)
		if err != nil {
			return err
		}

		// Assert the output to the expected type
		output, ok := r.(interface{})
		if !ok {
			return errors.New("failed to assert typed output to expected result type")
		}

		// Output the results
		utils.Outf("{{green}}output: {{/}} %+v\n", output)
	}
	return nil
}
