// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/requester"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"

	nconsts "github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	_ "github.com/nuklai/nuklaivm/registry" // ensure registry populated
)

type JSONRPCClient struct {
	requester *requester.EndpointRequester

	networkID uint32
	chainID   ids.ID
	g         *genesis.Genesis
	assetsL   sync.Mutex
	assets    map[string]*AssetReply
	assetsNFT map[string]*AssetNFTReply
	datasets  map[string]*DatasetReply
}

// New creates a new client object.
func NewJSONRPCClient(uri string, networkID uint32, chainID ids.ID) *JSONRPCClient {
	uri = strings.TrimSuffix(uri, "/")
	uri += JSONRPCEndpoint
	req := requester.New(uri, nconsts.Name)
	return &JSONRPCClient{
		requester: req,
		networkID: networkID,
		chainID:   chainID,
		assets:    map[string]*AssetReply{},
		assetsNFT: map[string]*AssetNFTReply{},
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

func (cli *JSONRPCClient) Tx(ctx context.Context, id ids.ID) (bool, bool, int64, uint64, string, error) {
	resp := new(TxReply)
	err := cli.requester.SendRequest(
		ctx,
		"tx",
		&TxArgs{TxID: id},
		resp,
	)
	switch {
	// We use string parsing here because the JSON-RPC library we use may not
	// allows us to perform errors.Is.
	case err != nil && strings.Contains(err.Error(), ErrTxNotFound.Error()):
		return false, false, -1, 0, "", nil
	case err != nil:
		return false, false, -1, 0, "", err
	}
	return true, resp.Success, resp.Timestamp, resp.Fee, resp.Actor, nil
}

func (cli *JSONRPCClient) Asset(
	ctx context.Context,
	asset string,
	useCache bool,
) (bool, string, string, string, uint8, string, string, uint64, uint64, string, string, string, string, string, error) {
	cli.assetsL.Lock()
	r, ok := cli.assets[asset]
	cli.assetsL.Unlock()
	if ok && useCache {
		return true, r.AssetType, r.Name, r.Symbol, r.Decimals, r.Metadata, r.URI, r.TotalSupply, r.MaxSupply, r.Admin, r.MintActor, r.PauseUnpauseActor, r.FreezeUnfreezeActor, r.EnableDisableKYCAccountActor, nil
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
	switch {
	// We use string parsing here because the JSON-RPC library we use may not
	// allows us to perform errors.Is.
	case err != nil && strings.Contains(err.Error(), ErrAssetNotFound.Error()):
		return false, "", "", "", 0, "", "", 0, 0, "", "", "", "", "", nil
	case err != nil:
		return false, "", "", "", 0, "", "", 0, 0, "", "", "", "", "", nil
	}
	cli.assetsL.Lock()
	cli.assets[asset] = resp
	cli.assetsL.Unlock()
	return true, resp.AssetType, resp.Name, resp.Symbol, resp.Decimals, resp.Metadata, resp.URI, resp.TotalSupply, resp.MaxSupply, resp.Admin, resp.MintActor, resp.PauseUnpauseActor, resp.FreezeUnfreezeActor, resp.EnableDisableKYCAccountActor, nil
}

func (cli *JSONRPCClient) AssetNFT(
	ctx context.Context,
	nft string,
	useCache bool,
) (bool, string, uint64, string, string, string, error) {
	cli.assetsL.Lock()
	r, ok := cli.assetsNFT[nft]
	cli.assetsL.Unlock()
	if ok && useCache {
		return true, r.CollectionID, r.UniqueID, r.URI, r.Metadata, r.Owner, nil
	}

	resp := new(AssetNFTReply)
	err := cli.requester.SendRequest(
		ctx,
		"assetNFT",
		&AssetArgs{
			Asset: nft,
		},
		resp,
	)
	switch {
	// We use string parsing here because the JSON-RPC library we use may not
	// allows us to perform errors.Is.
	case err != nil && strings.Contains(err.Error(), ErrAssetNotFound.Error()):
		return false, "", 0, "", "", "", nil
	case err != nil:
		return false, "", 0, "", "", "", nil
	}
	cli.assetsL.Lock()
	cli.assetsNFT[nft] = resp
	cli.assetsL.Unlock()
	return true, resp.CollectionID, resp.UniqueID, resp.URI, resp.Metadata, resp.Owner, nil
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

func (cli *JSONRPCClient) Dataset(
	ctx context.Context,
	dataset string,
	useCache bool,
) (bool, string, string, string, string, string, string, string, bool, string, string, uint64, uint8, uint8, uint8, uint8, string, error) {
	cli.assetsL.Lock()
	r, ok := cli.datasets[dataset]
	cli.assetsL.Unlock()
	if ok && useCache {
		return true, r.Name, r.Description, r.Categories, r.LicenseName, r.LicenseSymbol, r.LicenseURL, r.Metadata, r.IsCommunityDataset, r.SaleID, r.BaseAsset, r.BasePrice, r.RevenueModelDataShare, r.RevenueModelMetadataShare, r.RevenueModelDataOwnerCut, r.RevenueModelMetadataOwnerCut, r.Owner, nil
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
	switch {
	// We use string parsing here because the JSON-RPC library we use may not
	// allows us to perform errors.Is.
	case err != nil && strings.Contains(err.Error(), ErrDatasetNotFound.Error()):
		return false, "", "", "", "", "", "", "", false, "", "", 0, 0, 0, 0, 0, "", nil
	case err != nil:
		return false, "", "", "", "", "", "", "", false, "", "", 0, 0, 0, 0, 0, "", nil
	}
	cli.assetsL.Lock()
	cli.datasets[dataset] = resp
	cli.assetsL.Unlock()
	return true, resp.Name, resp.Description, resp.Categories, resp.LicenseName, resp.LicenseSymbol, resp.LicenseURL, resp.Metadata, resp.IsCommunityDataset, resp.SaleID, resp.BaseAsset, resp.BasePrice, resp.RevenueModelDataShare, resp.RevenueModelMetadataShare, resp.RevenueModelDataOwnerCut, resp.RevenueModelMetadataOwnerCut, resp.Owner, nil
}

func (cli *JSONRPCClient) DataContributionPending(ctx context.Context, dataset string) ([]DataContribution, error) {
	resp := new(DataContributionPendingReply)
	err := cli.requester.SendRequest(
		ctx,
		"dataContributionPending",
		&DatasetArgs{
			Dataset: dataset,
		},
		resp,
	)
	if err != nil {
		return []DataContribution{}, err
	}

	return resp.Contributions, nil
}

func (cli *JSONRPCClient) DatasetInfoFromMarketplace(ctx context.Context, dataset string) ([]DataContribution, error) {
	resp := new(DataContributionPendingReply)
	err := cli.requester.SendRequest(
		ctx,
		"datasetInfoFromMarketplace",
		&DatasetArgs{
			Dataset: dataset,
		},
		resp,
	)
	if err != nil {
		return []DataContribution{}, err
	}

	return resp.Contributions, nil
}

func (cli *JSONRPCClient) WaitForBalance(
	ctx context.Context,
	addr string,
	asset ids.ID,
	min uint64,
) error {
	exists, _, name, symbol, decimals, _, _, _, _, _, _, _, _, _, err := cli.Asset(ctx, asset.String(), true)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%s does not exist", asset)
	}

	return rpc.Wait(ctx, func(ctx context.Context) (bool, error) {
		balance, err := cli.Balance(ctx, addr, asset.String())
		if err != nil {
			return false, err
		}
		shouldExit := balance >= min
		if !shouldExit {
			utils.Outf(
				"{{yellow}}waiting for %s - %s %s on %s{{/}}\n",
				name,
				utils.FormatBalance(min, decimals),
				symbol,
				addr,
			)
		}
		return shouldExit, nil
	})
}

func (cli *JSONRPCClient) WaitForTransaction(ctx context.Context, txID ids.ID) (bool, uint64, string, error) {
	var success bool
	var fee uint64
	var actor string
	if err := rpc.Wait(ctx, func(ctx context.Context) (bool, error) {
		found, isuccess, _, ifee, iactor, err := cli.Tx(ctx, txID)
		if err != nil {
			return false, err
		}
		success = isuccess
		fee = ifee
		actor = iactor
		return found, nil
	}); err != nil {
		return false, 0, "", err
	}
	return success, fee, actor, nil
}

var _ chain.Parser = (*Parser)(nil)

type Parser struct {
	networkID uint32
	chainID   ids.ID
	genesis   *genesis.Genesis
}

func (p *Parser) Rules(t int64) chain.Rules {
	return p.genesis.Rules(t, p.networkID, p.chainID)
}

func (*Parser) Registry() (chain.ActionRegistry, chain.AuthRegistry) {
	return nconsts.ActionRegistry, nconsts.AuthRegistry
}

func (cli *JSONRPCClient) Parser(ctx context.Context) (chain.Parser, error) {
	g, err := cli.Genesis(ctx)
	if err != nil {
		return nil, err
	}
	return &Parser{cli.networkID, cli.chainID, g}, nil
}
