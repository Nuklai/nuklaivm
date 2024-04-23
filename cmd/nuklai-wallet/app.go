// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"runtime/debug"

	"github.com/nuklai/nuklaivm/cmd/nuklai-wallet/backend"

	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
	log logger.Logger
	b   *backend.Backend
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		log: logger.NewDefaultLogger(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.b = backend.New(func(err error) {
		a.log.Error(err.Error())
		runtime.Quit(ctx)
	})
	if err := a.b.Start(ctx); err != nil {
		a.log.Error(err.Error())
		runtime.Quit(ctx)
	}
}

// shutdown is called after the frontend is destroyed.
func (a *App) shutdown(ctx context.Context) {
	if err := a.b.Shutdown(ctx); err != nil {
		a.log.Error(err.Error())
	}
}

func (a *App) GetTotalBlocks() int {
	return a.b.GetTotalBlocks()
}

func (a *App) GetLatestBlocks(page int, count int) []*backend.BlockInfo {
	return a.b.GetLatestBlocks(page, count)
}

func (a *App) GetTransactionStats() []*backend.GenericInfo {
	return a.b.GetTransactionStats()
}

func (a *App) GetAccountStats() []*backend.GenericInfo {
	return a.b.GetAccountStats()
}

func (a *App) GetUnitPrices() []*backend.GenericInfo {
	return a.b.GetUnitPrices()
}

func (a *App) GetSubnetID() string {
	return a.b.GetSubnetID()
}

func (a *App) GetChainID() string {
	return a.b.GetChainID()
}

func (a *App) GetMyAssets() []*backend.AssetInfo {
	return a.b.GetMyAssets()
}

func (a *App) GetBalance() ([]*backend.BalanceInfo, error) {
	return a.b.GetBalance()
}

func (a *App) CreateAsset(symbol string, decimals string, metadata string) error {
	return a.b.CreateAsset(symbol, decimals, metadata)
}

func (a *App) MintAsset(asset string, address string, amount string) error {
	return a.b.MintAsset(asset, address, amount)
}

func (a *App) Transfer(asset string, address string, amount string, memo string) error {
	return a.b.Transfer(asset, address, amount, memo)
}

func (a *App) GetAddress() string {
	return a.b.GetAddress()
}

// TODO: Maybe find a different way to do this?
func (a *App) GetPrivateKey() string {
	return a.b.GetPrivateKey()
}

func (a *App) GetPublicKey() string {
	return a.b.GetPublicKey()
}

func (a *App) GetTransactions() *backend.Transactions {
	return a.b.GetTransactions()
}

func (a *App) StartFaucetSearch() (*backend.FaucetSearchInfo, error) {
	return a.b.StartFaucetSearch()
}

func (a *App) GetFaucetSolutions() *backend.FaucetSolutions {
	return a.b.GetFaucetSolutions()
}

func (a *App) GetAddressBook() []*backend.AddressInfo {
	return a.b.GetAddressBook()
}

func (a *App) AddAddressBook(name string, address string) error {
	return a.b.AddAddressBook(name, address)
}

func (a *App) GetAllAssets() []*backend.AssetInfo {
	return a.b.GetAllAssets()
}

func (a *App) AddAsset(asset string) error {
	return a.b.AddAsset(asset)
}

func (a *App) GetFeedInfo() (*backend.FeedInfo, error) {
	return a.b.GetFeedInfo()
}

func (a *App) GetFeed() ([]*backend.FeedObject, error) {
	return a.b.GetFeed(a.GetSubnetID(), a.GetChainID())
}

func (a *App) Message(message string, url string) error {
	return a.b.Message(message, url)
}

func (a *App) OpenLink(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}

func (*App) GetCommitHash() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}

func (a *App) UpdateNuklaiRPC(newNuklaiRPCUrl string) error {
	return a.b.UpdateNuklaiRPC(newNuklaiRPCUrl)
}

func (a *App) UpdateFaucetRPC(newFaucetRPCUrl string) {
	a.b.UpdateFaucetRPC(newFaucetRPCUrl)
}

func (a *App) UpdateFeedRPC(newFeedRPCUrl string) {
	a.b.UpdateFeedRPC(newFeedRPCUrl)
}

func (a *App) GetConfig() *backend.Config {
	return a.b.GetConfig()
}
