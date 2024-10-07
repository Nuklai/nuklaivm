// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/storage"

	"github.com/ava-labs/hypersdk/api/jsonrpc"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/requester"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"
)

const balanceCheckInterval = 500 * time.Millisecond

type JSONRPCClient struct {
	requester *requester.EndpointRequester
	g         *genesis.Genesis
	assetsL   sync.Mutex
	assets    map[string]*AssetReply
	datasets  map[string]*DatasetReply
}

// NewJSONRPCClient creates a new client object.
func NewJSONRPCClient(uri string) *JSONRPCClient {
	uri = strings.TrimSuffix(uri, "/")
	uri += JSONRPCEndpoint
	req := requester.New(uri, consts.Name)
	return &JSONRPCClient{
		requester: req,
		assets:    map[string]*AssetReply{},
		datasets:  map[string]*DatasetReply{},
	}
}

func (cli *JSONRPCClient) Genesis(ctx context.Context) (*genesis.Genesis, error) {
	if cli.g != nil {
		return cli.g, nil
	}

	resp := new(GenesisReply)
	err := cli.requester.SendRequest(
		ctx,
		"genesis",
		nil,
		resp,
	)
	if err != nil {
		return nil, err
	}
	cli.g = resp.Genesis
	return resp.Genesis, nil
}

func (cli *JSONRPCClient) Balance(ctx context.Context, addr string, asset string) (uint64, error) {
	resp := new(BalanceReply)
	err := cli.requester.SendRequest(
		ctx,
		"balance",
		&BalanceArgs{
			Address: addr,
			Asset:   asset,
		},
		resp,
	)
	return resp.Amount, err
}

func (cli *JSONRPCClient) Asset(
	ctx context.Context,
	asset string,
	useCache bool,
) (string, string, string, uint8, string, string, uint64, uint64, string, string, string, string, string, error) {
	cli.assetsL.Lock()
	r, ok := cli.assets[asset]
	cli.assetsL.Unlock()
	if ok && useCache {
		return r.AssetType, r.Name, r.Symbol, r.Decimals, r.Metadata, r.URI, r.TotalSupply, r.MaxSupply, r.Owner, r.MintAdmin, r.PauseUnpauseAdmin, r.FreezeUnfreezeAdmin, r.EnableDisableKYCAccountAdmin, nil
	}

	// Check if it's the native asset
	resp := new(AssetReply)
	err := cli.requester.SendRequest(
		ctx,
		"asset",
		&AssetArgs{
			Asset: asset,
		},
		resp,
	)
	if err != nil {
		return "", "", "", 0, "", "", 0, 0, "", "", "", "", "", err
	}
	cli.assetsL.Lock()
	cli.assets[asset] = resp
	cli.assetsL.Unlock()
	return resp.AssetType, resp.Name, resp.Symbol, resp.Decimals, resp.Metadata, resp.URI, resp.TotalSupply, resp.MaxSupply, resp.Owner, resp.MintAdmin, resp.PauseUnpauseAdmin, resp.FreezeUnfreezeAdmin, resp.EnableDisableKYCAccountAdmin, nil
}

func (cli *JSONRPCClient) Dataset(
	ctx context.Context,
	dataset string,
	useCache bool,
) (string, string, string, string, string, string, string, bool, string, string, uint64, uint8, uint8, uint8, uint8, string, error) {
	cli.assetsL.Lock()
	r, ok := cli.datasets[dataset]
	cli.assetsL.Unlock()
	if ok && useCache {
		return r.Name, r.Description, r.Categories, r.LicenseName, r.LicenseSymbol, r.LicenseURL, r.Metadata, r.IsCommunityDataset, r.MarketplaceAssetAddress, r.BaseAssetAddress, r.BasePrice, r.RevenueModelDataShare, r.RevenueModelMetadataShare, r.RevenueModelDataOwnerCut, r.RevenueModelMetadataOwnerCut, r.Owner, nil
	}

	resp := new(DatasetReply)
	err := cli.requester.SendRequest(
		ctx,
		"dataset",
		&DatasetArgs{
			Dataset: dataset,
		},
		resp,
	)
	if err != nil {
		return "", "", "", "", "", "", "", false, "", "", 0, 0, 0, 0, 0, "", nil
	}
	cli.assetsL.Lock()
	cli.datasets[dataset] = resp
	cli.assetsL.Unlock()
	return resp.Name, resp.Description, resp.Categories, resp.LicenseName, resp.LicenseSymbol, resp.LicenseURL, resp.Metadata, resp.IsCommunityDataset, resp.MarketplaceAssetAddress, resp.BaseAssetAddress, resp.BasePrice, resp.RevenueModelDataShare, resp.RevenueModelMetadataShare, resp.RevenueModelDataOwnerCut, resp.RevenueModelMetadataOwnerCut, resp.Owner, nil
}

func (cli *JSONRPCClient) DatasetContribution(ctx context.Context, dataset string) (string, string, string, string, bool, error) {
	resp := new(DatasetContributionReply)
	err := cli.requester.SendRequest(
		ctx,
		"datasetContribution",
		&DatasetArgs{
			Dataset: dataset,
		},
		resp,
	)
	if err != nil {
		return "", "", "", "", false, err
	}

	return resp.DatasetAddress, resp.DataLocation, resp.DataIdentifier, resp.Contributor, resp.Active, nil
}

func (cli *JSONRPCClient) EmissionInfo(ctx context.Context) (uint64, uint64, uint64, uint64, uint64, EmissionAccount, emission.EpochTracker, error) {
	resp := new(EmissionReply)
	err := cli.requester.SendRequest(
		ctx,
		"emissionInfo",
		nil,
		resp,
	)
	if err != nil {
		return 0, 0, 0, 0, 0, EmissionAccount{}, emission.EpochTracker{}, err
	}

	return resp.CurrentBlockHeight, resp.TotalSupply, resp.MaxSupply, resp.TotalStaked, resp.RewardsPerEpoch, resp.EmissionAccount, resp.EpochTracker, err
}

func (cli *JSONRPCClient) AllValidators(ctx context.Context) ([]*emission.Validator, error) {
	resp := new(ValidatorsReply)
	err := cli.requester.SendRequest(
		ctx,
		"allValidators",
		nil,
		resp,
	)
	if err != nil {
		return []*emission.Validator{}, err
	}
	return resp.Validators, err
}

func (cli *JSONRPCClient) StakedValidators(ctx context.Context) ([]*emission.Validator, error) {
	resp := new(ValidatorsReply)
	err := cli.requester.SendRequest(
		ctx,
		"stakedValidators",
		nil,
		resp,
	)
	if err != nil {
		return []*emission.Validator{}, err
	}
	return resp.Validators, err
}

func (cli *JSONRPCClient) ValidatorStake(ctx context.Context, nodeID ids.NodeID) (uint64, uint64, uint64, uint64, string, string, error) {
	resp := new(ValidatorStakeReply)
	err := cli.requester.SendRequest(
		ctx,
		"validatorStake",
		&ValidatorStakeArgs{
			NodeID: nodeID,
		},
		resp,
	)
	if err != nil {
		return 0, 0, 0, 0, "", "", err
	}
	return resp.StakeStartBlock, resp.StakeEndBlock, resp.StakedAmount, resp.DelegationFeeRate, resp.RewardAddress, resp.OwnerAddress, err
}

func (cli *JSONRPCClient) UserStake(ctx context.Context, owner string, nodeID string) (uint64, uint64, uint64, string, string, error) {
	resp := new(UserStakeReply)
	err := cli.requester.SendRequest(
		ctx,
		"userStake",
		&UserStakeArgs{
			Owner:  owner,
			NodeID: nodeID,
		},
		resp,
	)
	if err != nil {
		return 0, 0, 0, "", "", err
	}
	return resp.StakeStartBlock, resp.StakeEndBlock, resp.StakedAmount, resp.RewardAddress, resp.OwnerAddress, err
}

func (cli *JSONRPCClient) WaitForBalance(
	ctx context.Context,
	addr string,
	asset string,
	min uint64,
) error {
	return jsonrpc.Wait(ctx, balanceCheckInterval, func(ctx context.Context) (bool, error) {
		balance, err := cli.Balance(ctx, addr, asset)
		if err != nil {
			return false, err
		}
		shouldExit := balance >= min
		if !shouldExit {
			utils.Outf(
				"{{yellow}}waiting for %s balance: %s{{/}}\n",
				utils.FormatBalance(min),
				addr,
			)
		}
		return shouldExit, nil
	})
}

func (cli *JSONRPCClient) Parser(ctx context.Context) (chain.Parser, error) {
	g, err := cli.Genesis(ctx)
	if err != nil {
		return nil, err
	}
	return NewParser(g), nil
}

var _ chain.Parser = (*Parser)(nil)

type Parser struct {
	genesis *genesis.Genesis
}

func (p *Parser) Rules(_ int64) chain.Rules {
	return p.genesis.Rules
}

func (*Parser) ActionRegistry() chain.ActionRegistry {
	return ActionParser
}

func (*Parser) OutputRegistry() chain.OutputRegistry {
	return OutputParser
}

func (*Parser) AuthRegistry() chain.AuthRegistry {
	return AuthParser
}

func (*Parser) StateManager() chain.StateManager {
	return &storage.StateManager{}
}

func NewParser(genesis *genesis.Genesis) chain.Parser {
	return &Parser{genesis: genesis}
}

// Used as a lambda function for creating ExternalSubscriberServer parser
func CreateParser(genesisBytes []byte) (chain.Parser, error) {
	var genesis genesis.Genesis
	if err := json.Unmarshal(genesisBytes, &genesis); err != nil {
		return nil, err
	}
	return NewParser(&genesis), nil
}

func (cli *JSONRPCClient) Simulate(ctx context.Context, callTx actions.ContractCall, actor codec.Address) (state.Keys, uint64, error) {
	resp := new(SimulateCallTxReply)
	err := cli.requester.SendRequest(
		ctx,
		"simulateCallContractTx",
		&SimulateCallTxArgs{CallTx: callTx, Actor: actor},
		resp,
	)
	if err != nil {
		return nil, 0, err
	}
	result := state.Keys{}
	for _, entry := range resp.StateKeys {
		hexBytes, err := hex.DecodeString(entry.HexKey)
		if err != nil {
			return nil, 0, err
		}

		result.Add(string(hexBytes), state.Permissions(entry.Permissions))
	}
	return result, resp.FuelConsumed, nil
}
