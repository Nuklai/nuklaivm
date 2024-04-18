// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package backend

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/pebble"
	hutils "github.com/ava-labs/hypersdk/utils"
	tconsts "github.com/nuklai/nuklaivm/consts"
)

const (
	keyPrefix         = 0x0
	assetsPrefix      = 0x1
	transactionPrefix = 0x2
	searchPrefix      = 0x3
	addressPrefix     = 0x4
)

type Storage struct {
	db database.Database
}

func OpenStorage(databasePath string) (*Storage, error) {
	db, _, err := pebble.New(databasePath, pebble.NewDefaultConfig())
	if err != nil {
		return nil, err
	}
	return &Storage{db}, nil
}

// generateKey generates a key for storage with prefixes for chain and type.
func (*Storage) generateKey(prefix byte, subnetID, chainID ids.ID, keyData []byte) []byte {
	subnetPrefix := subnetID[:]
	chainPrefix := chainID[:]
	k := make([]byte, 1+len(subnetPrefix)+len(chainPrefix)+len(keyData))
	k[0] = prefix
	copy(k[1:], subnetPrefix)
	copy(k[1+len(subnetPrefix):], chainPrefix)
	copy(k[1+len(subnetPrefix)+len(chainPrefix):], keyData)
	return k
}

func (s *Storage) StoreKey(privateKey ed25519.PrivateKey) error {
	has, err := s.db.Has([]byte{keyPrefix})
	if err != nil {
		return err
	}
	if has {
		return ErrDuplicate
	}
	return s.db.Put([]byte{keyPrefix}, privateKey[:])
}

func (s *Storage) GetKey() (ed25519.PrivateKey, error) {
	v, err := s.db.Get([]byte{keyPrefix})
	if errors.Is(err, database.ErrNotFound) {
		return ed25519.EmptyPrivateKey, nil
	}
	if err != nil {
		return ed25519.EmptyPrivateKey, err
	}
	return ed25519.PrivateKey(v), nil
}

func (s *Storage) StoreAsset(subnetID, chainID ids.ID, assetID ids.ID, owned bool) error {
	k := s.generateKey(assetsPrefix, subnetID, chainID, assetID[:])
	v := []byte{0x0}
	if owned {
		v = []byte{0x1}
	}
	return s.db.Put(k, v)
}

func (s *Storage) HasAsset(subnetID, chainID ids.ID, assetID ids.ID) (bool, error) {
	k := s.generateKey(assetsPrefix, subnetID, chainID, assetID[:])
	return s.db.Has(k)
}

func (s *Storage) GetAssets(subnetID, chainID ids.ID) ([]ids.ID, []bool, error) {
	prefix := s.generateKey(assetsPrefix, subnetID, chainID, nil)
	iter := s.db.NewIteratorWithPrefix(prefix)
	defer iter.Release()

	assets := []ids.ID{}
	owned := []bool{}
	for iter.Next() {
		id, _ := ids.ToID(iter.Key()[len(prefix):])
		assets = append(assets, id)
		owned = append(owned, iter.Value()[0] == 0x1)
	}
	return assets, owned, iter.Error()
}

func (s *Storage) StoreTransaction(subnetID, chainID ids.ID, tx *TransactionInfo) error {
	txID, err := ids.FromString(tx.ID)
	if err != nil {
		return err
	}
	inverseTime := consts.MaxUint64 - uint64(time.Now().UnixMilli())
	txKey := append(binary.BigEndian.AppendUint64(nil, inverseTime), txID[:]...)
	k := s.generateKey(transactionPrefix, subnetID, chainID, txKey)
	b, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return s.db.Put(k, b)
}

func (s *Storage) GetTransactions(subnetID, chainID ids.ID) ([]*TransactionInfo, error) {
	prefix := s.generateKey(transactionPrefix, subnetID, chainID, nil)
	iter := s.db.NewIteratorWithPrefix(prefix)
	defer iter.Release()

	txs := []*TransactionInfo{}
	for iter.Next() {
		var tx TransactionInfo
		if err := json.Unmarshal(iter.Value(), &tx); err != nil {
			return nil, err
		}
		txs = append(txs, &tx)
	}
	return txs, iter.Error()
}

func (s *Storage) StoreAddress(subnetID, chainID ids.ID, address string, nickname string) error {
	addr, err := codec.ParseAddressBech32(tconsts.HRP, address)
	if err != nil {
		return err
	}
	k := s.generateKey(addressPrefix, subnetID, chainID, addr[:])
	return s.db.Put(k, []byte(nickname))
}

func (s *Storage) GetAddresses(subnetID, chainID ids.ID) ([]*AddressInfo, error) {
	prefix := s.generateKey(addressPrefix, subnetID, chainID, nil)
	iter := s.db.NewIteratorWithPrefix(prefix)
	defer iter.Release()

	addresses := []*AddressInfo{}
	for iter.Next() {
		address := codec.Address(iter.Key()[len(prefix):])
		nickname := string(iter.Value())
		addresses = append(addresses, &AddressInfo{nickname, codec.MustAddressBech32(tconsts.HRP, address), fmt.Sprintf("%s [%s..%s]", nickname, address[:len(tconsts.HRP)+3], address[len(address)-3:])})
	}
	return addresses, iter.Error()
}

func (s *Storage) StoreSolution(subnetID, chainID ids.ID, solution *FaucetSearchInfo) error {
	solutionID := hutils.ToID([]byte(solution.Solution))
	inverseTime := consts.MaxUint64 - uint64(time.Now().UnixMilli())
	solutionKey := append(binary.BigEndian.AppendUint64(nil, inverseTime), solutionID[:]...)
	k := s.generateKey(searchPrefix, subnetID, chainID, solutionKey)
	b, err := json.Marshal(solution)
	if err != nil {
		return err
	}
	return s.db.Put(k, b)
}

func (s *Storage) GetSolutions(subnetID, chainID ids.ID) ([]*FaucetSearchInfo, error) {
	prefix := s.generateKey(searchPrefix, subnetID, chainID, nil)
	iter := s.db.NewIteratorWithPrefix(prefix)
	defer iter.Release()

	solutions := []*FaucetSearchInfo{}
	for iter.Next() {
		var solution FaucetSearchInfo
		if err := json.Unmarshal(iter.Value(), &solution); err != nil {
			return nil, err
		}
		solutions = append(solutions, &solution)
	}
	return solutions, iter.Error()
}

func (s *Storage) DeleteDBKey(subnetID, chainID ids.ID, keyData []byte) error {
	// Determine the full key using the known structure of the stored data
	k := s.generateKey(keyData[0], subnetID, chainID, keyData[1:])
	return s.db.Delete(k)
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("unable to close database: %w", err)
	}
	return nil
}
