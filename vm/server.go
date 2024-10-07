// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"encoding/hex"
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/emission"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/utils"

	"github.com/ava-labs/hypersdk/api"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/x/contracts/runtime"
)

const JSONRPCEndpoint = "/nuklaiapi"

var _ api.HandlerFactory[api.VM] = (*jsonRPCServerFactory)(nil)

type jsonRPCServerFactory struct{}

func (jsonRPCServerFactory) New(v api.VM) (api.Handler, error) {
	handler, err := api.NewJSONRPCHandler(consts.Name, NewJSONRPCServer(v))
	return api.Handler{
		Path:    JSONRPCEndpoint,
		Handler: handler,
	}, err
}

type JSONRPCServer struct {
	vm api.VM
}

func NewJSONRPCServer(vm api.VM) *JSONRPCServer {
	return &JSONRPCServer{vm: vm}
}

type GenesisReply struct {
	Genesis *genesis.Genesis `json:"genesis"`
}

func (j *JSONRPCServer) Genesis(_ *http.Request, _ *struct{}, reply *GenesisReply) (err error) {
	reply.Genesis = j.vm.Genesis().(*genesis.Genesis)
	return nil
}

type BalanceArgs struct {
	Address string `json:"address"`
	Asset   string `json:"asset"`
}

type BalanceReply struct {
	Amount uint64 `json:"amount"`
}

func (j *JSONRPCServer) Balance(req *http.Request, args *BalanceArgs, reply *BalanceReply) error {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.Balance")
	defer span.End()

	addr, err := codec.StringToAddress(args.Address)
	if err != nil {
		return err
	}
	assetAddr, err := utils.GetAssetAddressBySymbol(args.Asset)
	if err != nil {
		return err
	}
	balance, err := storage.GetAssetAccountBalanceFromState(ctx, j.vm.ReadState, assetAddr, addr)
	if err != nil {
		return err
	}
	reply.Amount = balance

	return nil
}

type AssetArgs struct {
	Asset string `json:"asset"`
}

type AssetReply struct {
	AssetType                    string `json:"assetType"`
	Name                         string `json:"name"`
	Symbol                       string `json:"symbol"`
	Decimals                     uint8  `json:"decimals"`
	Metadata                     string `json:"metadata"`
	URI                          string `json:"uri"`
	TotalSupply                  uint64 `json:"totalSupply"`
	MaxSupply                    uint64 `json:"maxSupply"`
	Owner                        string `json:"owner"`
	MintAdmin                    string `json:"mintAdmin"`
	PauseUnpauseAdmin            string `json:"pauseUnpauseAdmin"`
	FreezeUnfreezeAdmin          string `json:"freezeUnfreezeAdmin"`
	EnableDisableKYCAccountAdmin string `json:"enableDisableKYCAccountAdmin"`
}

func (j *JSONRPCServer) Asset(req *http.Request, args *AssetArgs, reply *AssetReply) error {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.Asset")
	defer span.End()

	assetAddress, err := utils.GetAssetAddressBySymbol(args.Asset)
	if err != nil {
		return err
	}
	assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, owner, mintAdmin, pauseUnpauseAdmin, freezeUnfreezeAdmin, enableDisableKYCAccountAdmin, err := storage.GetAssetInfoFromState(ctx, j.vm.ReadState, assetAddress)
	if err != nil {
		return err
	}
	switch assetType {
	case consts.AssetFungibleTokenID:
		reply.AssetType = consts.AssetFungibleTokenDesc
	case consts.AssetNonFungibleTokenID:
		reply.AssetType = consts.AssetNonFungibleTokenDesc
	case consts.AssetFractionalTokenID:
		reply.AssetType = consts.AssetFractionalTokenDesc
	case consts.AssetMarketplaceTokenID:
		reply.AssetType = consts.AssetMarketplaceTokenDesc
	}
	reply.Name = string(name)
	reply.Symbol = string(symbol)
	reply.Decimals = decimals
	reply.Metadata = string(metadata)
	reply.URI = string(uri)
	reply.TotalSupply = totalSupply
	reply.MaxSupply = maxSupply
	reply.Owner = owner.String()
	reply.MintAdmin = mintAdmin.String()
	reply.PauseUnpauseAdmin = pauseUnpauseAdmin.String()
	reply.FreezeUnfreezeAdmin = freezeUnfreezeAdmin.String()
	reply.EnableDisableKYCAccountAdmin = enableDisableKYCAccountAdmin.String()

	return nil
}

type DatasetArgs struct {
	Dataset string `json:"dataset"`
}

type DatasetReply struct {
	Name                         string `json:"name"`
	Description                  string `json:"description"`
	Categories                   string `json:"categories"`
	LicenseName                  string `json:"licenseName"`
	LicenseSymbol                string `json:"licenseSymbol"`
	LicenseURL                   string `json:"licenseURL"`
	Metadata                     string `json:"metadata"`
	IsCommunityDataset           bool   `json:"isCommunityDataset"`
	MarketplaceAssetAddress      string `json:"marketplaceAssetAddress"`
	BaseAssetAddress             string `json:"baseAssetAddress"`
	BasePrice                    uint64 `json:"basePrice"`
	RevenueModelDataShare        uint8  `json:"revenueModelDataShare"`
	RevenueModelMetadataShare    uint8  `json:"revenueModelMetadataShare"`
	RevenueModelDataOwnerCut     uint8  `json:"revenueModelDataOwnerCut"`
	RevenueModelMetadataOwnerCut uint8  `json:"revenueModelMetadataOwnerCut"`
	Owner                        string `json:"owner"`
}

func (j *JSONRPCServer) Dataset(req *http.Request, args *DatasetArgs, reply *DatasetReply) error {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.Dataset")
	defer span.End()

	datasetAddress, err := utils.GetAssetAddressBySymbol(args.Dataset)
	if err != nil {
		return err
	}
	name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, marketplaceAssetAddress, baseAssetAddress, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadatOwnerCut, owner, err := storage.GetDatasetInfoFromState(ctx, j.vm.ReadState, datasetAddress)
	if err != nil {
		return err
	}
	reply.Name = string(name)
	reply.Description = string(description)
	reply.Categories = string(categories)
	reply.LicenseName = string(licenseName)
	reply.LicenseSymbol = string(licenseSymbol)
	reply.LicenseURL = string(licenseURL)
	reply.Metadata = string(metadata)
	reply.IsCommunityDataset = isCommunityDataset
	reply.MarketplaceAssetAddress = marketplaceAssetAddress.String()
	reply.BaseAssetAddress = baseAssetAddress.String()
	reply.BasePrice = basePrice
	reply.RevenueModelDataShare = revenueModelDataShare
	reply.RevenueModelMetadataShare = revenueModelMetadataShare
	reply.RevenueModelDataOwnerCut = revenueModelDataOwnerCut
	reply.RevenueModelMetadataOwnerCut = revenueModelMetadatOwnerCut
	reply.Owner = owner.String()

	return nil
}

type DatasetContributionArgs struct {
	ContributionID string `json:"contributionID"`
}

type DatasetContributionReply struct {
	DatasetAddress string `json:"datasetAddress"`
	DataLocation   string `json:"dataLocation"`
	DataIdentifier string `json:"dataIdentifier"`
	Contributor    string `json:"contributor"`
	Active         bool   `json:"active"`
}

func (j *JSONRPCServer) DataContributionPending(req *http.Request, args *DatasetContributionArgs, reply *DatasetContributionReply) (err error) {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.DatasetContribution")
	defer span.End()

	contributionID, err := ids.FromString(args.ContributionID)
	if err != nil {
		return err
	}

	datasetAddress, dataLocation, dataIdentifier, contributor, active, err := storage.GetDatasetContributionInfoFromState(ctx, j.vm.ReadState, contributionID)
	if err != nil {
		return err
	}
	reply.DatasetAddress = datasetAddress.String()
	reply.DataLocation = string(dataLocation)
	reply.DataIdentifier = string(dataIdentifier)
	reply.Contributor = contributor.String()
	reply.Active = active

	return nil
}

type EmissionAccount struct {
	Address           string `json:"address"`
	AccumulatedReward uint64 `json:"accumulatedReward"`
}

type EmissionReply struct {
	CurrentBlockHeight uint64                `json:"currentBlockHeight"`
	TotalSupply        uint64                `json:"totalSupply"`
	MaxSupply          uint64                `json:"maxSupply"`
	TotalStaked        uint64                `json:"totalStaked"`
	RewardsPerEpoch    uint64                `json:"rewardsPerEpoch"`
	EmissionAccount    EmissionAccount       `json:"emissionAccount"`
	EpochTracker       emission.EpochTracker `json:"epochTracker"`
}

func (j *JSONRPCServer) EmissionInfo(req *http.Request, _ *struct{}, reply *EmissionReply) (err error) {
	_, span := j.vm.Tracer().Start(req.Context(), "Server.EmissionInfo")
	defer span.End()

	emissionAccount, totalSupply, maxSupply, totalStaked, epochTracker := emissionTracker.GetInfo()

	currentBlockHeight, totalSupply, maxSupply, totalStaked, rewardsPerEpoch, emissionAccount, epochTracker := emissionTracker.GetLastAcceptedBlockHeight(), totalSupply, maxSupply, totalStaked, emissionTracker.GetRewardsPerEpoch(), emissionAccount, epochTracker
	reply.CurrentBlockHeight = currentBlockHeight
	reply.TotalSupply = totalSupply
	reply.MaxSupply = maxSupply
	reply.TotalStaked = totalStaked
	reply.RewardsPerEpoch = rewardsPerEpoch
	reply.EmissionAccount.Address = emissionAccount.Address.String()
	reply.EmissionAccount.AccumulatedReward = emissionAccount.AccumulatedReward
	reply.EpochTracker = epochTracker
	return nil
}

type ValidatorsReply struct {
	Validators []*emission.Validator `json:"validators"`
}

func (j *JSONRPCServer) AllValidators(req *http.Request, _ *struct{}, reply *ValidatorsReply) (err error) {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.AllValidators")
	defer span.End()

	validators := emissionTracker.GetAllValidators(ctx)
	reply.Validators = validators
	return nil
}

func (j *JSONRPCServer) StakedValidators(req *http.Request, _ *struct{}, reply *ValidatorsReply) (err error) {
	_, span := j.vm.Tracer().Start(req.Context(), "Server.StakedValidators")
	defer span.End()

	validators := emissionTracker.GetStakedValidator(ids.EmptyNodeID)
	reply.Validators = validators
	return nil
}

type ValidatorStakeArgs struct {
	NodeID ids.NodeID `json:"nodeID"`
}

type ValidatorStakeReply struct {
	StakeStartBlock   uint64 `json:"stakeStartBlock"`   // Start block of the stake
	StakeEndBlock     uint64 `json:"stakeEndBlock"`     // End block of the stake
	StakedAmount      uint64 `json:"stakedAmount"`      // Amount of NAI staked
	DelegationFeeRate uint64 `json:"delegationFeeRate"` // Delegation fee rate
	RewardAddress     string `json:"rewardAddress"`     // Address to receive rewards
	OwnerAddress      string `json:"ownerAddress"`      // Address of the owner who registered the validator
}

func (j *JSONRPCServer) ValidatorStake(req *http.Request, args *ValidatorStakeArgs, reply *ValidatorStakeReply) (err error) {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.ValidatorStake")
	defer span.End()

	exists, stakeStartBlock, stakeEndBlock, stakedAmount, delegationFeeRate, rewardAddress, ownerAddress, err := storage.GetValidatorStakeFromState(ctx, j.vm.ReadState, args.NodeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrValidatorStakeNotFound
	}

	reply.StakeStartBlock = stakeStartBlock
	reply.StakeEndBlock = stakeEndBlock
	reply.StakedAmount = stakedAmount
	reply.DelegationFeeRate = delegationFeeRate
	reply.RewardAddress = rewardAddress.String()
	reply.OwnerAddress = ownerAddress.String()
	return nil
}

type UserStakeArgs struct {
	Owner  string `json:"owner"`
	NodeID string `json:"nodeID"`
}

type UserStakeReply struct {
	StakeStartBlock uint64 `json:"stakeStartBlock"` // Start block of the stake
	StakeEndBlock   uint64 `json:"stakeEndBlock"`   // End block of the stake
	StakedAmount    uint64 `json:"stakedAmount"`    // Amount of NAI staked
	RewardAddress   string `json:"rewardAddress"`   // Address to receive rewards
	OwnerAddress    string `json:"ownerAddress"`    // Address of the owner who delegated
}

func (j *JSONRPCServer) UserStake(req *http.Request, args *UserStakeArgs, reply *UserStakeReply) (err error) {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.UserStake")
	defer span.End()

	ownerID, err := codec.StringToAddress(args.Owner)
	if err != nil {
		return err
	}
	nodeID, err := ids.NodeIDFromString(args.NodeID)
	if err != nil {
		return err
	}

	exists, stakeStartBlock, stakeEndBlock, stakedAmount, rewardAddress, ownerAddress, err := storage.GetDelegatorStakeFromState(ctx, j.vm.ReadState, ownerID, nodeID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDelegatorStakeNotFound
	}

	reply.StakeStartBlock = stakeStartBlock
	reply.StakeEndBlock = stakeEndBlock
	reply.StakedAmount = stakedAmount
	reply.RewardAddress = rewardAddress.String()
	reply.OwnerAddress = ownerAddress.String()
	return nil
}

type SimulateCallTxArgs struct {
	CallTx actions.ContractCall `json:"callTx"`
	Actor  codec.Address        `json:"actor"`
}

type SimulateStateKey struct {
	HexKey      string `json:"hex"`
	Permissions byte   `json:"perm"`
}
type SimulateCallTxReply struct {
	StateKeys    []SimulateStateKey `json:"stateKeys"`
	FuelConsumed uint64             `json:"fuel"`
}

func (j *JSONRPCServer) SimulateCallContractTx(req *http.Request, args *SimulateCallTxArgs, reply *SimulateCallTxReply) (err error) {
	stateKeys, fuelConsumed, err := j.simulate(req.Context(), args.CallTx, args.Actor)
	if err != nil {
		return err
	}
	reply.StateKeys = make([]SimulateStateKey, 0, len(stateKeys))
	for key, permission := range stateKeys {
		reply.StateKeys = append(reply.StateKeys, SimulateStateKey{HexKey: hex.EncodeToString([]byte(key)), Permissions: byte(permission)})
	}
	reply.FuelConsumed = fuelConsumed
	return nil
}

func (j *JSONRPCServer) simulate(ctx context.Context, t actions.ContractCall, actor codec.Address) (state.Keys, uint64, error) {
	currentState, err := j.vm.ImmutableState(ctx)
	if err != nil {
		return nil, 0, err
	}
	recorder := storage.NewRecorder(currentState)
	startFuel := uint64(1000000000)
	callInfo := &runtime.CallInfo{
		Contract:     t.ContractAddress,
		Actor:        actor,
		State:        &storage.ContractStateManager{Mutable: recorder},
		FunctionName: t.Function,
		Params:       t.CallData,
		Fuel:         startFuel,
		Value:        t.Value,
	}
	_, err = wasmRuntime.CallContract(ctx, callInfo)
	return recorder.GetStateKeys(), startFuel - callInfo.RemainingFuel(), err
}
