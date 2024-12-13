// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/vm"
	"gopkg.in/yaml.v2"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/utils"
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
		utils.Outf("{{green}}fee consumed:{{/}} %s NAI\n", utils.FormatBalance(result.Fee))

		// Use NewReader to create a Packer from the result output
		packer := codec.NewReader(result.Outputs[0], len(result.Outputs[0]))
		r, err := vm.OutputParser.Unmarshal(packer)
		if err != nil {
			return err
		}

		// Automatically handle all types by reflecting and marshaling to JSON
		reflectValue := reflect.ValueOf(r)
		if !reflectValue.IsValid() || reflectValue.IsZero() {
			return errors.New("result is invalid or nil")
		}

		// Marshal the result to JSON for a generic and readable output
		outputJSON, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		utils.Outf("{{green}}output:{{/}} %s\n", string(outputJSON))
	}
	return nil
}
