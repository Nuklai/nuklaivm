// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/consts"
	"github.com/nuklai/nuklaivm/genesis"
	"github.com/nuklai/nuklaivm/storage"

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
	assetID, err := getAssetIDBySymbol(args.Asset)
	if err != nil {
		return err
	}
	balance, err := storage.GetBalanceFromState(ctx, j.vm.ReadState, addr, assetID)
	if err != nil {
		return err
	}
	reply.Amount = balance
	return err
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
	Admin                        string `json:"admin"`
	MintActor                    string `json:"mintActor"`
	PauseUnpauseActor            string `json:"pauseUnpauseActor"`
	FreezeUnfreezeActor          string `json:"freezeUnfreezeActor"`
	EnableDisableKYCAccountActor string `json:"enableDisableKYCAccountActor"`
}

func (j *JSONRPCServer) Asset(req *http.Request, args *AssetArgs, reply *AssetReply) error {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.Asset")
	defer span.End()

	assetID, err := getAssetIDBySymbol(args.Asset)
	if err != nil {
		return err
	}
	exists, assetType, name, symbol, decimals, metadata, uri, totalSupply, maxSupply, admin, mintActor, pauseUnpauseActor, freezeUnfreezeActor, enableDisableKYCAccountActor, err := storage.GetAssetFromState(ctx, j.vm.ReadState, assetID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAssetNotFound
	}
	switch assetType {
	case consts.AssetFungibleTokenID:
		reply.AssetType = consts.AssetFungibleTokenDesc
	case consts.AssetNonFungibleTokenID:
		reply.AssetType = consts.AssetNonFungibleTokenDesc
	case consts.AssetDatasetTokenID:
		reply.AssetType = consts.AssetDatasetTokenDesc
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
	reply.Admin = admin.String()
	reply.MintActor = mintActor.String()
	reply.PauseUnpauseActor = pauseUnpauseActor.String()
	reply.FreezeUnfreezeActor = freezeUnfreezeActor.String()
	reply.EnableDisableKYCAccountActor = enableDisableKYCAccountActor.String()
	return err
}

type AssetNFTReply struct {
	CollectionID string `json:"collectionID"`
	UniqueID     uint64 `json:"uniqueID"`
	URI          string `json:"uri"`
	Metadata     string `json:"metadata"`
	Owner        string `json:"owner"`
}

func (j *JSONRPCServer) AssetNFT(req *http.Request, args *AssetArgs, reply *AssetNFTReply) error {
	ctx, span := j.vm.Tracer().Start(req.Context(), "Server.AssetNFT")
	defer span.End()

	nftID, err := getAssetIDBySymbol(args.Asset)
	if err != nil {
		return err
	}
	exists, collectionID, uniqueID, uri, metadata, owner, err := storage.GetAssetNFTFromState(ctx, j.vm.ReadState, nftID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAssetNotFound
	}

	reply.CollectionID = collectionID.String()
	reply.UniqueID = uniqueID
	reply.URI = string(uri)
	reply.Metadata = string(metadata)
	reply.Owner = owner.String()

	return err
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
	SaleID                       string `json:"saleID"`
	BaseAsset                    string `json:"baseAsset"`
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

	datasetID, err := getAssetIDBySymbol(args.Dataset)
	if err != nil {
		return err
	}
	exists, name, description, categories, licenseName, licenseSymbol, licenseURL, metadata, isCommunityDataset, saleID, baseAsset, basePrice, revenueModelDataShare, revenueModelMetadataShare, revenueModelDataOwnerCut, revenueModelMetadatOwnerCut, owner, err := storage.GetDatasetFromState(ctx, j.vm.ReadState, datasetID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrDatasetNotFound
	}
	reply.Name = string(name)
	reply.Description = string(description)
	reply.Categories = string(categories)
	reply.LicenseName = string(licenseName)
	reply.LicenseSymbol = string(licenseSymbol)
	reply.LicenseURL = string(licenseURL)
	reply.Metadata = string(metadata)
	reply.IsCommunityDataset = isCommunityDataset
	reply.SaleID = saleID.String()
	reply.BaseAsset = baseAsset.String()
	reply.BasePrice = basePrice
	reply.RevenueModelDataShare = revenueModelDataShare
	reply.RevenueModelMetadataShare = revenueModelMetadataShare
	reply.RevenueModelDataOwnerCut = revenueModelDataOwnerCut
	reply.RevenueModelMetadataOwnerCut = revenueModelMetadatOwnerCut
	reply.Owner = owner.String()
	return err
}

type DataContribution struct {
	DataLocation   string `json:"dataLocation"`
	DataIdentifier string `json:"dataIdentifier"`
	Contributor    string `json:"contributor"`
}

type DataContributionPendingReply struct {
	Contributions []DataContribution `json:"contributions"`
}

func (j *JSONRPCServer) DataContributionPending(req *http.Request, args *DatasetArgs, reply *DataContributionPendingReply) (err error) {
	/*
		ctx, span := j.vm.Tracer().Start(req.Context(), "Server.DataContributionPending")
		defer span.End()

		datasetID, err := getAssetIDBySymbol(args.Dataset)
		if err != nil {
			return err
		}

		// Get all contributions for the dataset
		 	contributions, err := j.vm.GetDataContributionPending(ctx, datasetID, codec.EmptyAddress)
		   	if err != nil {
		   		return err
		   	}

		   	// Iterate over contributions and populate reply
		   	for _, contrib := range contributions {
		   		convertedContribution := DataContribution{
		   			DataLocation:   string(contrib.DataLocation),                              // Convert []byte to string
		   			DataIdentifier: string(contrib.DataIdentifier),                            // Convert []byte to string
		   			Contributor:    codec.MustAddressBech32(nconsts.HRP, contrib.Contributor), // Convert codec.Address to string
		   		}
		   		reply.Contributions = append(reply.Contributions, convertedContribution)
		   	}
	*/

	_, span := j.vm.Tracer().Start(req.Context(), "Server.DataContributionPending")
	defer span.End()

	_, err = getAssetIDBySymbol(args.Dataset)
	if err != nil {
		return err
	}
	reply.Contributions = []DataContribution{}
	return nil
}

func getAssetIDBySymbol(symbol string) (ids.ID, error) {
	if strings.TrimSpace(symbol) == "" || strings.EqualFold(symbol, consts.Symbol) {
		return ids.Empty, nil
	}
	return ids.FromString(symbol)
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