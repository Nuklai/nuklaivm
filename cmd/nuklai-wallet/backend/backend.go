// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package backend

import (
	"container/list"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/utils/units"
	"github.com/ava-labs/hypersdk/chain"
	hcli "github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/pubsub"
	"github.com/ava-labs/hypersdk/rpc"
	hutils "github.com/ava-labs/hypersdk/utils"
	"github.com/ava-labs/hypersdk/window"
	"github.com/nuklai/nuklaivm/actions"
	"github.com/nuklai/nuklaivm/auth"
	"github.com/nuklai/nuklaivm/challenge"
	frpc "github.com/nuklai/nuklaivm/cmd/nuklai-faucet/rpc"
	"github.com/nuklai/nuklaivm/cmd/nuklai-feed/manager"
	ferpc "github.com/nuklai/nuklaivm/cmd/nuklai-feed/rpc"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

// Constant for max blocks to store
const maxBlocksStore = 250 // Adjust as needed

var (
	databaseFolder string
	configFile     string
)

type Backend struct {
	ctx   context.Context
	fatal func(error)

	s *Storage
	c *Config

	priv    ed25519.PrivateKey
	factory *auth.ED25519Factory
	addr    codec.Address
	addrStr string

	cli      *rpc.JSONRPCClient
	subnetID ids.ID
	chainID  ids.ID
	scli     *rpc.WebSocketClient
	ncli     *nrpc.JSONRPCClient
	parser   chain.Parser
	fcli     *frpc.JSONRPCClient
	fecli    *ferpc.JSONRPCClient

	blockLock   sync.Mutex
	blocks      *list.List               // Change slice to a linked list for deque operations
	blockMap    map[uint64]*list.Element // Map block heights to elements in the list
	stats       []*TimeStat
	currentStat *TimeStat

	txAlertLock       sync.Mutex
	transactionAlerts []*Alert

	searchLock   sync.Mutex
	search       *FaucetSearchInfo
	searchAlerts []*Alert

	htmlCache *cache.LRU[string, *HTMLMeta]
	urlQueue  chan string
	stopChans map[string]chan bool // Map to hold channels that control the stopping of goroutines
}

// NewApp creates a new App application struct
func New(fatal func(error)) *Backend {
	return &Backend{
		fatal:             fatal,
		blocks:            list.New(),
		blockMap:          make(map[uint64]*list.Element),
		stats:             []*TimeStat{},
		transactionAlerts: []*Alert{},
		searchAlerts:      []*Alert{},
		htmlCache:         &cache.LRU[string, *HTMLMeta]{Size: 128},
		urlQueue:          make(chan string, 128),
	}
}

func (b *Backend) Start(ctx context.Context) error {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default paths")
	}
	// Set default values to the current directory
	defaultDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory: " + err.Error())
	}

	// Read environment variables or use default values
	databaseFolder = getEnv("NUKLAI_WALLET_DB_PATH", filepath.Join(defaultDir, ".nuklai-wallet/db"))
	configFile = getEnv("NUKLAI_WALLET_CONFIG_PATH", filepath.Join(defaultDir, ".nuklai-wallet/config.json"))

	if err := os.MkdirAll(databaseFolder, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create database directory '%s': %w", databaseFolder, err)
	}

	// Set context
	b.ctx = ctx

	// Open storage
	s, err := OpenStorage(databaseFolder)
	if err != nil {
		return err
	}
	b.s = s

	// Generate key
	key, err := s.GetKey()
	if err != nil {
		return err
	}
	if key == ed25519.EmptyPrivateKey {
		// TODO: encrypt key
		priv, err := ed25519.GeneratePrivateKey()
		if err != nil {
			return err
		}
		if err := s.StoreKey(priv); err != nil {
			return err
		}
		key = priv
	}
	b.priv = key
	b.factory = auth.NewED25519Factory(b.priv)
	b.addr = auth.NewED25519Address(b.priv.PublicKey())
	b.addrStr = codec.MustAddressBech32(nconsts.HRP, b.addr)

	if err := b.AddAddressBook("Me", b.addrStr); err != nil {
		return err
	}
	if err := b.s.StoreAsset(b.subnetID, b.chainID, ids.Empty, false); err != nil {
		return err
	}

	if err := b.initClients(); err != nil {
		return err
	}

	b.stopChans = make(map[string]chan bool)
	b.stopChans["collectBlocks"] = make(chan bool)
	b.stopChans["parseURLs"] = make(chan bool)

	// Start fetching blocks
	go b.collectBlocks()
	go b.parseURLs()
	return nil
}

func (b *Backend) initClients() error {
	// Open config
	rawConfig, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("Failed to read config file '%s', using default configuration: %v", configFile, err)
		// TODO: replace with DEVNET
		b.c = &Config{ // Default configuration, should be configurable
			NuklaiRPC:   "http://localhost:9090",
			FaucetRPC:   "http://localhost:9091",
			SearchCores: 4,
			FeedRPC:     "http://localhost:9092",
		}
	} else {
		if err := json.Unmarshal(rawConfig, &b.c); err != nil {
			return fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Create clients
	b.cli = rpc.NewJSONRPCClient(b.c.NuklaiRPC)
	networkID, subnetID, chainID, err := b.cli.Network(b.ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch network info: %w", err)
	}
	b.subnetID = subnetID
	b.chainID = chainID

	b.scli, err = rpc.NewWebSocketClient(b.c.NuklaiRPC, rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
	if err != nil {
		return fmt.Errorf("failed to initialize WebSocket client: %w", err)
	}

	b.ncli = nrpc.NewJSONRPCClient(b.c.NuklaiRPC, networkID, chainID)
	if b.parser, err = b.ncli.Parser(b.ctx); err != nil {
		return fmt.Errorf("failed to initialize parser: %w", err)
	}

	b.fcli = frpc.NewJSONRPCClient(b.c.FaucetRPC)
	b.fecli = ferpc.NewJSONRPCClient(b.c.FeedRPC)
	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func (b *Backend) restartCollectBlocks() {
	select {
	case b.stopChans["collectBlocks"] <- true:
		// Signal to stop the existing goroutine
	default:
		// No need to block if it's already stopped
	}

	go b.collectBlocks()
}

func (b *Backend) collectBlocks() {
	for {
		select {
		case <-b.stopChans["collectBlocks"]:
			return // Stop the goroutine
		default:
			if err := b.scli.RegisterBlocks(); err != nil {
				b.fatal(err)
				return
			}

			var (
				start     time.Time
				lastBlock int64
				tpsWindow = window.Window{}
			)
			for b.ctx.Err() == nil {
				blk, results, prices, err := b.scli.ListenBlock(b.ctx, b.parser)
				if err != nil {
					b.fatal(err)
					return
				}
				consumed := chain.Dimensions{}
				failTxs := 0
				for i, result := range results {
					nconsumed, err := chain.Add(consumed, result.Consumed)
					if err != nil {
						b.fatal(err)
						return
					}
					consumed = nconsumed

					tx := blk.Txs[i]
					actor := tx.Auth.Actor()
					if !result.Success {
						failTxs++
					}

					// We should exit action parsing as soon as possible
					switch action := tx.Action.(type) {
					case *actions.Transfer:
						if actor != b.addr && action.To != b.addr {
							continue
						}
						_, symbol, decimals, _, _, owner, _, err := b.ncli.Asset(b.ctx, action.Asset, true)
						if err != nil {
							b.fatal(err)
							return
						}
						txInfo := &TransactionInfo{
							ID:        tx.ID().String(),
							Size:      fmt.Sprintf("%.2fKB", float64(tx.Size())/units.KiB),
							Success:   result.Success,
							Timestamp: blk.Tmstmp,
							Actor:     codec.MustAddressBech32(nconsts.HRP, actor),
							Type:      "Transfer",
							Units:     hcli.ParseDimensions(result.Consumed),
							Fee:       fmt.Sprintf("%s %s", hutils.FormatBalance(result.Fee, nconsts.Decimals), nconsts.Symbol),
						}
						if result.Success {
							txInfo.Summary = fmt.Sprintf("%s %s -> %s", hutils.FormatBalance(action.Value, decimals), symbol, codec.MustAddressBech32(nconsts.HRP, action.To))
							if len(action.Memo) > 0 {
								txInfo.Summary += fmt.Sprintf(" (memo: %s)", action.Memo)
							}
						} else {
							txInfo.Summary = string(result.Output)
						}
						if action.To == b.addr {
							if actor != b.addr && result.Success {
								b.txAlertLock.Lock()
								b.transactionAlerts = append(b.transactionAlerts, &Alert{"info", fmt.Sprintf("Received %s %s from Transfer", hutils.FormatBalance(action.Value, decimals), symbol)})
								b.txAlertLock.Unlock()
							}
							hasAsset, err := b.s.HasAsset(b.subnetID, b.chainID, action.Asset)
							if err != nil {
								b.fatal(err)
								return
							}
							if !hasAsset {
								if err := b.s.StoreAsset(b.subnetID, b.chainID, action.Asset, b.addrStr == owner); err != nil {
									b.fatal(err)
									return
								}
							}
							if err := b.s.StoreTransaction(b.subnetID, b.chainID, txInfo); err != nil {
								b.fatal(err)
								return
							}
						} else if actor == b.addr {
							if err := b.s.StoreTransaction(b.subnetID, b.chainID, txInfo); err != nil {
								b.fatal(err)
								return
							}
						}
					case *actions.CreateAsset:
						if actor != b.addr {
							continue
						}
						if err := b.s.StoreAsset(b.subnetID, b.chainID, tx.ID(), true); err != nil {
							b.fatal(err)
							return
						}
						txInfo := &TransactionInfo{
							ID:        tx.ID().String(),
							Size:      fmt.Sprintf("%.2fKB", float64(tx.Size())/units.KiB),
							Success:   result.Success,
							Timestamp: blk.Tmstmp,
							Actor:     codec.MustAddressBech32(nconsts.HRP, actor),
							Type:      "CreateAsset",
							Units:     hcli.ParseDimensions(result.Consumed),
							Fee:       fmt.Sprintf("%s %s", hutils.FormatBalance(result.Fee, nconsts.Decimals), nconsts.Symbol),
						}
						if result.Success {
							txInfo.Summary = fmt.Sprintf("assetID: %s symbol: %s decimals: %d metadata: %s", tx.ID(), action.Symbol, action.Decimals, action.Metadata)
						} else {
							txInfo.Summary = string(result.Output)
						}
						if err := b.s.StoreTransaction(b.subnetID, b.chainID, txInfo); err != nil {
							b.fatal(err)
							return
						}
					case *actions.MintAsset:
						if actor != b.addr && action.To != b.addr {
							continue
						}
						_, symbol, decimals, _, _, owner, _, err := b.ncli.Asset(b.ctx, action.Asset, true)
						if err != nil {
							b.fatal(err)
							return
						}
						txInfo := &TransactionInfo{
							ID:        tx.ID().String(),
							Timestamp: blk.Tmstmp,
							Size:      fmt.Sprintf("%.2fKB", float64(tx.Size())/units.KiB),
							Success:   result.Success,
							Actor:     codec.MustAddressBech32(nconsts.HRP, actor),
							Type:      "Mint",
							Units:     hcli.ParseDimensions(result.Consumed),
							Fee:       fmt.Sprintf("%s %s", hutils.FormatBalance(result.Fee, nconsts.Decimals), nconsts.Symbol),
						}
						if result.Success {
							txInfo.Summary = fmt.Sprintf("%s %s -> %s", hutils.FormatBalance(action.Value, decimals), symbol, codec.MustAddressBech32(nconsts.HRP, action.To))
						} else {
							txInfo.Summary = string(result.Output)
						}
						if action.To == b.addr {
							if actor != b.addr && result.Success {
								b.txAlertLock.Lock()
								b.transactionAlerts = append(b.transactionAlerts, &Alert{"info", fmt.Sprintf("Received %s %s from Mint", hutils.FormatBalance(action.Value, decimals), symbol)})
								b.txAlertLock.Unlock()
							}
							hasAsset, err := b.s.HasAsset(b.subnetID, b.chainID, action.Asset)
							if err != nil {
								b.fatal(err)
								return
							}
							if !hasAsset {
								if err := b.s.StoreAsset(b.subnetID, b.chainID, action.Asset, b.addrStr == owner); err != nil {
									b.fatal(err)
									return
								}
							}
							if err := b.s.StoreTransaction(b.subnetID, b.chainID, txInfo); err != nil {
								b.fatal(err)
								return
							}
						} else if actor == b.addr {
							if err := b.s.StoreTransaction(b.subnetID, b.chainID, txInfo); err != nil {
								b.fatal(err)
								return
							}
						}
					}
				}

				now := time.Now()
				if start.IsZero() {
					start = now
				}
				bi := &BlockInfo{}
				if lastBlock != 0 {
					since := now.Unix() - lastBlock
					newWindow, err := window.Roll(tpsWindow, int(since))
					if err != nil {
						b.fatal(err)
						return
					}
					tpsWindow = newWindow
					window.Update(&tpsWindow, window.WindowSliceSize-consts.Uint64Len, uint64(len(blk.Txs)))
					runningDuration := time.Since(start)
					tpsDivisor := math.Min(window.WindowSize, runningDuration.Seconds())
					bi.TPS = fmt.Sprintf("%.2f", float64(window.Sum(tpsWindow))/tpsDivisor)
					bi.Latency = time.Now().UnixMilli() - blk.Tmstmp
				} else {
					window.Update(&tpsWindow, window.WindowSliceSize-consts.Uint64Len, uint64(len(blk.Txs)))
					bi.TPS = "0.0"
				}
				blkID, err := blk.ID()
				if err != nil {
					b.fatal(err)
					return
				}
				bi.Timestamp = blk.Tmstmp
				bi.ID = blkID.String()
				bi.Height = blk.Hght
				bi.Size = fmt.Sprintf("%.2fKB", float64(blk.Size())/units.KiB)
				bi.Consumed = hcli.ParseDimensions(consumed)
				bi.Prices = hcli.ParseDimensions(prices)
				bi.StateRoot = blk.StateRoot.String()
				bi.FailTxs = failTxs
				bi.Txs = len(blk.Txs)

				b.AddBlock(bi) // Manage block storage with AddBlock

				// Statistical updates
				b.blockLock.Lock()
				sTime := blk.Tmstmp / consts.MillisecondsPerSecond
				if b.currentStat != nil && b.currentStat.Timestamp != sTime {
					b.stats = append(b.stats, b.currentStat)
					b.currentStat = nil
				}
				if b.currentStat == nil {
					b.currentStat = &TimeStat{Timestamp: sTime, Accounts: set.Set[string]{}}
				}
				b.currentStat.Transactions += bi.Txs
				for _, tx := range blk.Txs {
					b.currentStat.Accounts.Add(codec.MustAddressBech32(nconsts.HRP, tx.Auth.Sponsor()))
				}
				b.currentStat.Prices = prices
				snow := time.Now().Unix()
				newStart := 0
				for i, item := range b.stats {
					newStart = i
					if snow-item.Timestamp < 120 {
						break
					}
				}
				b.stats = b.stats[newStart:]
				b.blockLock.Unlock()

				lastBlock = now.Unix()
			}
		}
	}
}

func (b *Backend) Shutdown(context.Context) error {
	_ = b.scli.Close()
	return b.s.Close()
}

func (b *Backend) GetTotalBlocks() int {
	b.blockLock.Lock()
	defer b.blockLock.Unlock()

	return b.blocks.Len()
}

func (b *Backend) GetLatestBlocks(page int, count int) []*BlockInfo {
	b.blockLock.Lock()
	defer b.blockLock.Unlock()

	startIndex := (page - 1) * count
	endIndex := startIndex + count

	// Collect blocks for the page
	var blocks []*BlockInfo
	current := b.blocks.Front()
	for i := 0; current != nil && len(blocks) < endIndex; i++ {
		if i >= startIndex {
			blocks = append(blocks, current.Value.(*BlockInfo))
		}
		current = current.Next()
	}

	// Return only the slice of blocks for the current page
	if len(blocks) > count {
		blocks = blocks[:count]
	}
	return blocks
}

// Method to add a new block to the store, ensuring the store size is managed
func (b *Backend) AddBlock(newBlock *BlockInfo) {
	b.blockLock.Lock()
	defer b.blockLock.Unlock()

	// Prepend new block to the deque (front of the list)
	element := b.blocks.PushFront(newBlock)
	b.blockMap[newBlock.Height] = element

	// Trim the oldest blocks if the deque size exceeds maxBlocksStore
	if b.blocks.Len() > maxBlocksStore {
		lastElement := b.blocks.Back()
		b.blocks.Remove(lastElement)
		delete(b.blockMap, lastElement.Value.(*BlockInfo).Height)
	}
}

func (b *Backend) GetTransactionStats() []*GenericInfo {
	b.blockLock.Lock()
	defer b.blockLock.Unlock()

	info := make([]*GenericInfo, len(b.stats))
	for i := 0; i < len(b.stats); i++ {
		info[i] = &GenericInfo{b.stats[i].Timestamp, uint64(b.stats[i].Transactions), ""}
	}
	return info
}

func (b *Backend) GetAccountStats() []*GenericInfo {
	b.blockLock.Lock()
	defer b.blockLock.Unlock()

	info := make([]*GenericInfo, len(b.stats))
	for i := 0; i < len(b.stats); i++ {
		info[i] = &GenericInfo{b.stats[i].Timestamp, uint64(b.stats[i].Accounts.Len()), ""}
	}
	return info
}

func (b *Backend) GetUnitPrices() []*GenericInfo {
	b.blockLock.Lock()
	defer b.blockLock.Unlock()

	info := make([]*GenericInfo, 0, len(b.stats)*chain.FeeDimensions)
	for i := 0; i < len(b.stats); i++ {
		info = append(info, &GenericInfo{b.stats[i].Timestamp, b.stats[i].Prices[0], "Bandwidth"})
		info = append(info, &GenericInfo{b.stats[i].Timestamp, b.stats[i].Prices[1], "Compute"})
		info = append(info, &GenericInfo{b.stats[i].Timestamp, b.stats[i].Prices[2], "Storage [Read]"})
		info = append(info, &GenericInfo{b.stats[i].Timestamp, b.stats[i].Prices[3], "Storage [Allocate]"})
		info = append(info, &GenericInfo{b.stats[i].Timestamp, b.stats[i].Prices[4], "Storage [Write]"})
	}
	return info
}

func (b *Backend) GetSubnetID() string {
	return b.subnetID.String()
}

func (b *Backend) GetChainID() string {
	return b.chainID.String()
}

func (b *Backend) GetMyAssets() []*AssetInfo {
	assets := []*AssetInfo{}
	assetIDs, owned, err := b.s.GetAssets(b.subnetID, b.chainID)
	if err != nil {
		b.fatal(err)
		return nil
	}
	for i, asset := range assetIDs {
		if !owned[i] {
			continue
		}
		_, symbol, decimals, metadata, supply, owner, _, err := b.ncli.Asset(b.ctx, asset, false)
		if err != nil {
			b.fatal(err)
			return nil
		}
		strAsset := asset.String()
		assets = append(assets, &AssetInfo{
			ID:        asset.String(),
			Symbol:    string(symbol),
			Decimals:  int(decimals),
			Metadata:  string(metadata),
			Supply:    hutils.FormatBalance(supply, decimals),
			Creator:   owner,
			StrSymbol: fmt.Sprintf("%s [%s..%s]", symbol, strAsset[:3], strAsset[len(strAsset)-3:]),
		})
	}
	return assets
}

func (b *Backend) CreateAsset(symbol string, decimals string, metadata string) error {
	// Ensure have sufficient balance
	bal, err := b.ncli.Balance(b.ctx, b.addrStr, ids.Empty)
	if err != nil {
		return err
	}

	// Generate transaction
	udecimals, err := strconv.ParseUint(decimals, 10, 8)
	if err != nil {
		return err
	}
	_, tx, maxFee, err := b.cli.GenerateTransaction(b.ctx, b.parser, nil, &actions.CreateAsset{
		Symbol:   []byte(symbol),
		Decimals: uint8(udecimals),
		Metadata: []byte(metadata),
	}, b.factory)
	if err != nil {
		return fmt.Errorf("%w: unable to generate transaction", err)
	}
	if maxFee > bal {
		return fmt.Errorf("insufficient balance (have: %s %s, want: %s %s)", hutils.FormatBalance(bal, nconsts.Decimals), nconsts.Symbol, hutils.FormatBalance(maxFee, nconsts.Decimals), nconsts.Symbol)
	}
	if err := b.scli.RegisterTx(tx); err != nil {
		return err
	}

	// Wait for transaction
	_, dErr, result, err := b.scli.ListenTx(b.ctx)
	if err != nil {
		return err
	}
	if dErr != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("transaction failed on-chain: %s", result.Output)
	}
	return nil
}

func (b *Backend) MintAsset(asset string, address string, amount string) error {
	// Input validation
	assetID, err := ids.FromString(asset)
	if err != nil {
		return err
	}
	_, _, decimals, _, _, _, _, err := b.ncli.Asset(b.ctx, assetID, true)
	if err != nil {
		return err
	}
	value, err := hutils.ParseBalance(amount, decimals)
	if err != nil {
		return err
	}
	to, err := codec.ParseAddressBech32(nconsts.HRP, address)
	if err != nil {
		return err
	}

	// Ensure have sufficient balance
	bal, err := b.ncli.Balance(b.ctx, b.addrStr, ids.Empty)
	if err != nil {
		return err
	}

	// Generate transaction
	_, tx, maxFee, err := b.cli.GenerateTransaction(b.ctx, b.parser, nil, &actions.MintAsset{
		To:    to,
		Asset: assetID,
		Value: value,
	}, b.factory)
	if err != nil {
		return fmt.Errorf("%w: unable to generate transaction", err)
	}
	if maxFee > bal {
		return fmt.Errorf("insufficient balance (have: %s %s, want: %s %s)", hutils.FormatBalance(bal, nconsts.Decimals), nconsts.Symbol, hutils.FormatBalance(maxFee, nconsts.Decimals), nconsts.Symbol)
	}
	if err := b.scli.RegisterTx(tx); err != nil {
		return err
	}

	// Wait for transaction
	_, dErr, result, err := b.scli.ListenTx(b.ctx)
	if err != nil {
		return err
	}
	if dErr != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("transaction failed on-chain: %s", result.Output)
	}
	return nil
}

func (b *Backend) Transfer(asset string, address string, amount string, memo string) error {
	// Input validation
	assetID, err := ids.FromString(asset)
	if err != nil {
		return err
	}
	_, symbol, decimals, _, _, _, _, err := b.ncli.Asset(b.ctx, assetID, true)
	if err != nil {
		return err
	}
	value, err := hutils.ParseBalance(amount, decimals)
	if err != nil {
		return err
	}
	to, err := codec.ParseAddressBech32(nconsts.HRP, address)
	if err != nil {
		return err
	}

	// Ensure have sufficient balance for transfer
	sendBal, err := b.ncli.Balance(b.ctx, b.addrStr, assetID)
	if err != nil {
		return err
	}
	if value > sendBal {
		return fmt.Errorf("insufficient balance (have: %s %s, want: %s %s)", hutils.FormatBalance(sendBal, decimals), symbol, hutils.FormatBalance(value, decimals), symbol)
	}

	// Ensure have sufficient balance for fees
	bal, err := b.ncli.Balance(b.ctx, b.addrStr, ids.Empty)
	if err != nil {
		return err
	}

	// Generate transaction
	_, tx, maxFee, err := b.cli.GenerateTransaction(b.ctx, b.parser, nil, &actions.Transfer{
		To:    to,
		Asset: assetID,
		Value: value,
		Memo:  []byte(memo),
	}, b.factory)
	if err != nil {
		return fmt.Errorf("%w: unable to generate transaction", err)
	}
	if assetID != ids.Empty {
		if maxFee > bal {
			return fmt.Errorf("insufficient balance (have: %s %s, want: %s %s)", hutils.FormatBalance(bal, nconsts.Decimals), nconsts.Symbol, hutils.FormatBalance(maxFee, nconsts.Decimals), nconsts.Symbol)
		}
	} else {
		if maxFee+value > bal {
			return fmt.Errorf("insufficient balance (have: %s %s, want: %s %s)", hutils.FormatBalance(bal, nconsts.Decimals), nconsts.Symbol, hutils.FormatBalance(maxFee+value, nconsts.Decimals), nconsts.Symbol)
		}
	}
	if err := b.scli.RegisterTx(tx); err != nil {
		return err
	}

	// Wait for transaction
	_, dErr, result, err := b.scli.ListenTx(b.ctx)
	if err != nil {
		return err
	}
	if dErr != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("transaction failed on-chain: %s", result.Output)
	}
	return nil
}

func (b *Backend) GetAddress() string {
	return b.addrStr
}

func (b *Backend) GetPrivateKey() string {
	return hex.EncodeToString(b.priv[:])
}

func (b *Backend) GetPublicKey() string {
	pubKey := b.priv.PublicKey()
	return hex.EncodeToString(pubKey[:])
}

func (b *Backend) GetBalance() ([]*BalanceInfo, error) {
	assets, _, err := b.s.GetAssets(b.subnetID, b.chainID)
	if err != nil {
		return nil, err
	}
	balances := []*BalanceInfo{}
	for _, asset := range assets {
		_, symbol, decimals, _, _, _, _, err := b.ncli.Asset(b.ctx, asset, true)
		if err != nil {
			return nil, err
		}
		bal, err := b.ncli.Balance(b.ctx, b.addrStr, asset)
		if err != nil {
			return nil, err
		}
		strAsset := asset.String()
		if asset == ids.Empty {
			balances = append(balances, &BalanceInfo{ID: asset.String(), Str: fmt.Sprintf("%s %s", hutils.FormatBalance(bal, decimals), symbol), Bal: fmt.Sprintf("%s (Balance: %s)", symbol, hutils.FormatBalance(bal, decimals)), Has: bal > 0})
		} else {
			balances = append(balances, &BalanceInfo{ID: asset.String(), Str: fmt.Sprintf("%s %s [%s]", hutils.FormatBalance(bal, decimals), symbol, asset), Bal: fmt.Sprintf("%s [%s..%s] (Balance: %s)", symbol, strAsset[:3], strAsset[len(strAsset)-3:], hutils.FormatBalance(bal, decimals)), Has: bal > 0})
		}
	}
	return balances, nil
}

func (b *Backend) GetTransactions() *Transactions {
	b.txAlertLock.Lock()
	defer b.txAlertLock.Unlock()

	var alerts []*Alert
	if len(b.transactionAlerts) > 0 {
		alerts = b.transactionAlerts
		b.transactionAlerts = []*Alert{}
	}
	txs, err := b.s.GetTransactions(b.subnetID, b.chainID)
	if err != nil {
		b.fatal(err)
		return nil
	}
	return &Transactions{alerts, txs}
}

func (b *Backend) StartFaucetSearch() (*FaucetSearchInfo, error) {
	b.searchLock.Lock()
	if b.search != nil {
		b.searchLock.Unlock()
		return nil, errors.New("already searching")
	}
	b.search = &FaucetSearchInfo{}
	b.searchLock.Unlock()

	address, err := b.fcli.FaucetAddress(b.ctx)
	if err != nil {
		b.searchLock.Lock()
		b.search = nil
		b.searchLock.Unlock()
		return nil, err
	}
	salt, difficulty, err := b.fcli.Challenge(b.ctx)
	if err != nil {
		b.searchLock.Lock()
		b.search = nil
		b.searchLock.Unlock()
		return nil, err
	}
	b.search.FaucetAddress = address
	b.search.Salt = hex.EncodeToString(salt)
	b.search.Difficulty = difficulty

	// Search in the background
	go func() {
		start := time.Now()
		solution, attempts := challenge.Search(salt, difficulty, b.c.SearchCores)
		txID, amount, err := b.fcli.SolveChallenge(b.ctx, b.addrStr, salt, solution)
		b.searchLock.Lock()
		b.search.Solution = hex.EncodeToString(solution)
		b.search.Attempts = attempts
		b.search.Elapsed = time.Since(start).String()
		if err == nil {
			b.search.TxID = txID.String()
			b.search.Amount = fmt.Sprintf("%s %s", hutils.FormatBalance(amount, nconsts.Decimals), nconsts.Symbol)
			b.searchAlerts = append(b.searchAlerts, &Alert{"success", fmt.Sprintf("Search Successful [Attempts: %d, Elapsed: %s]", attempts, b.search.Elapsed)})
		} else {
			b.search.Err = err.Error()
			b.searchAlerts = append(b.searchAlerts, &Alert{"error", fmt.Sprintf("Search Failed: %v", err)})
		}
		search := b.search
		b.search = nil
		b.searchLock.Unlock()
		if err := b.s.StoreSolution(b.subnetID, b.chainID, search); err != nil {
			b.fatal(err)
		}
	}()
	return b.search, nil
}

func (b *Backend) GetFaucetSolutions() *FaucetSolutions {
	solutions, err := b.s.GetSolutions(b.subnetID, b.chainID)
	if err != nil {
		b.fatal(err)
		return nil
	}

	b.searchLock.Lock()
	defer b.searchLock.Unlock()

	var alerts []*Alert
	if len(b.searchAlerts) > 0 {
		alerts = b.searchAlerts
		b.searchAlerts = []*Alert{}
	}

	return &FaucetSolutions{alerts, b.search, solutions}
}

func (b *Backend) GetAddressBook() []*AddressInfo {
	addresses, err := b.s.GetAddresses(b.subnetID, b.chainID)
	if err != nil {
		b.fatal(err)
		return nil
	}
	return addresses
}

// Any existing address will be overwritten with a new name
func (b *Backend) AddAddressBook(name string, address string) error {
	name = strings.TrimSpace(name)
	address = strings.TrimSpace(address)
	return b.s.StoreAddress(b.subnetID, b.chainID, address, name)
}

func (b *Backend) GetAllAssets() []*AssetInfo {
	arr, _, err := b.s.GetAssets(b.subnetID, b.chainID)
	if err != nil {
		b.fatal(err)
		return nil
	}
	assets := []*AssetInfo{}
	for _, asset := range arr {
		_, symbol, decimals, metadata, supply, owner, _, err := b.ncli.Asset(b.ctx, asset, false)
		if err != nil {
			b.fatal(err)
			return nil
		}
		strAsset := asset.String()
		assets = append(assets, &AssetInfo{
			ID:        asset.String(),
			Symbol:    string(symbol),
			Decimals:  int(decimals),
			Metadata:  string(metadata),
			Supply:    hutils.FormatBalance(supply, decimals),
			Creator:   owner,
			StrSymbol: fmt.Sprintf("%s [%s..%s]", symbol, strAsset[:3], strAsset[len(strAsset)-3:]),
		})
	}
	return assets
}

func (b *Backend) AddAsset(asset string) error {
	assetID, err := ids.FromString(asset)
	if err != nil {
		return err
	}
	hasAsset, err := b.s.HasAsset(b.subnetID, b.chainID, assetID)
	if err != nil {
		return err
	}
	if hasAsset {
		return nil
	}
	exists, _, _, _, _, owner, _, err := b.ncli.Asset(b.ctx, assetID, true)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAssetMissing
	}
	return b.s.StoreAsset(b.subnetID, b.chainID, assetID, owner == b.addrStr)
}

func (b *Backend) GetFeedInfo() (*FeedInfo, error) {
	addr, fee, err := b.fecli.FeedInfo(context.TODO())
	if err != nil {
		return nil, err
	}
	return &FeedInfo{
		addr,
		fmt.Sprintf("%s %s", hutils.FormatBalance(fee, nconsts.Decimals), nconsts.Symbol),
	}, nil
}

func (b *Backend) restartParseURLs() {
	select {
	case b.stopChans["parseURLs"] <- true:
		// Signal to stop the existing goroutine
	default:
		// No need to block if it's already stopped
	}

	go b.parseURLs()
}

func (b *Backend) parseURLs() {
	for {
		select {
		case <-b.stopChans["parseURLs"]:
			return // Stop the goroutine
		default:
			client := http.DefaultClient
			for {
				select {
				case u := <-b.urlQueue:
					// Protect against maliciously crafted URLs
					parsedURL, err := url.Parse(u)
					if err != nil {
						continue
					}
					if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
						continue
					}
					ip := net.ParseIP(parsedURL.Host)
					if ip != nil {
						if ip.IsPrivate() || ip.IsLoopback() {
							continue
						}
					}

					// Attempt to fetch URL contents
					ctx, cancel := context.WithTimeout(b.ctx, 30*time.Second)
					req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
					if err != nil {
						cancel()
						continue
					}
					resp, err := client.Do(req)
					if err != nil {
						fmt.Println("unable to fetch URL", err)
						// We already put the URL in as nil in
						// our cache, so we won't refetch it.
						cancel()
						continue
					}
					b.htmlCache.Put(u, ParseHTML(u, parsedURL.Host, resp.Body))
					_ = resp.Body.Close()
					cancel()
				case <-b.ctx.Done():
					return
				}
			}
		}
	}
}

func (b *Backend) GetFeed(subnetID, chainID string) ([]*FeedObject, error) {
	feed, err := b.fecli.Feed(context.TODO(), subnetID, chainID)
	if err != nil {
		return nil, err
	}
	nfeed := make([]*FeedObject, 0, len(feed))
	for _, fo := range feed {
		tfo := &FeedObject{
			SubnetID:  fo.SubnetID,
			ChainID:   fo.ChainID,
			Address:   fo.Address,
			ID:        fo.TxID.String(),
			Timestamp: fo.Timestamp,
			Fee:       fmt.Sprintf("%s %s", hutils.FormatBalance(fo.Fee, nconsts.Decimals), nconsts.Symbol),

			Message: fo.Content.Message,
			URL:     fo.Content.URL,
		}
		if len(fo.Content.URL) > 0 {
			if m, ok := b.htmlCache.Get(fo.Content.URL); ok {
				tfo.URLMeta = m
			} else {
				b.htmlCache.Put(fo.Content.URL, nil) // ensure we don't refetch
				b.urlQueue <- fo.Content.URL
			}
		}
		nfeed = append(nfeed, tfo)
	}
	return nfeed, nil
}

func (b *Backend) Message(message string, url string) error {
	// Get latest feed info
	recipient, fee, err := b.fecli.FeedInfo(context.TODO())
	if err != nil {
		return err
	}
	recipientAddr, err := codec.ParseAddressBech32(nconsts.HRP, recipient)
	if err != nil {
		return err
	}

	// Encode data
	fc := &manager.FeedContent{
		Message: message,
		URL:     url,
	}
	data, err := json.Marshal(fc)
	if err != nil {
		return err
	}

	// Ensure have sufficient balance
	bal, err := b.ncli.Balance(b.ctx, b.addrStr, ids.Empty)
	if err != nil {
		return err
	}

	// Generate transaction
	_, tx, maxFee, err := b.cli.GenerateTransaction(b.ctx, b.parser, nil, &actions.Transfer{
		To:    recipientAddr,
		Asset: ids.Empty,
		Value: fee,
		Memo:  data,
	}, b.factory)
	if err != nil {
		return fmt.Errorf("%w: unable to generate transaction", err)
	}
	if maxFee+fee > bal {
		return fmt.Errorf("insufficient balance (have: %s %s, want: %s %s)", hutils.FormatBalance(bal, nconsts.Decimals), nconsts.Symbol, hutils.FormatBalance(maxFee+fee, nconsts.Decimals), nconsts.Symbol)
	}
	if err := b.scli.RegisterTx(tx); err != nil {
		return err
	}

	// Wait for transaction
	_, dErr, result, err := b.scli.ListenTx(b.ctx)
	if err != nil {
		return err
	}
	if dErr != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("transaction failed on-chain: %s", result.Output)
	}
	return nil
}

// UpdateNuklaiRPC updates the RPC url for Nuklai
func (b *Backend) UpdateNuklaiRPC(newNuklaiRPCUrl string) error {
	b.c.NuklaiRPC = newNuklaiRPCUrl

	// Reinitialize the JSONRPCClient with the new URL
	newClient := rpc.NewJSONRPCClient(b.c.NuklaiRPC)

	// Refresh the network information
	networkID, subnetID, chainID, err := newClient.Network(b.ctx)
	if err != nil {
		return fmt.Errorf("failed to make network call with new client: %w", err)
	}
	// If successful, replace the old client
	b.cli = newClient
	b.subnetID = subnetID
	b.chainID = chainID

	// Also update for websocket client if needed
	b.scli, err = rpc.NewWebSocketClient(newNuklaiRPCUrl, rpc.DefaultHandshakeTimeout, pubsub.MaxPendingMessages, pubsub.MaxReadMessageSize)
	if err != nil {
		return fmt.Errorf("failed to reinitialize WebSocketClient: %w", err)
	}

	b.ncli = nrpc.NewJSONRPCClient(newNuklaiRPCUrl, networkID, chainID)
	parser, err := b.ncli.Parser(b.ctx)
	if err != nil {
		return err
	}
	b.parser = parser

	// Update clients as well
	_, err = b.fcli.UpdateNuklaiRPC(b.ctx, newNuklaiRPCUrl)
	if err != nil {
		return err
	}
	_, err = b.fecli.UpdateNuklaiRPC(b.ctx, newNuklaiRPCUrl)
	if err != nil {
		return err
	}

	// Restart the goroutines
	b.restartCollectBlocks()
	b.restartParseURLs()

	return nil
}

func (b *Backend) UpdateFaucetRPC(newFaucetRPCUrl string) {
	b.c.FaucetRPC = newFaucetRPCUrl
	b.fcli = frpc.NewJSONRPCClient(b.c.FaucetRPC)
}

func (b *Backend) UpdateFeedRPC(newFeedRPCUrl string) {
	b.c.FeedRPC = newFeedRPCUrl
	b.fecli = ferpc.NewJSONRPCClient(b.c.FeedRPC)
}

func (b *Backend) GetConfig() *Config {
	return b.c
}
