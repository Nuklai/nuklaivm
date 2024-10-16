// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
)

var _ (chain.StateManager) = (*StateManager)(nil)

type StateManager struct{}

func (*StateManager) CanDeduct(
	ctx context.Context,
	addr codec.Address,
	im state.Immutable,
	amount uint64,
) error {
	bal, err := GetAssetAccountBalanceNoController(ctx, im, NAIAddress, addr)
	if err != nil {
		return err
	}
	if bal < amount {
		return ErrInvalidBalance
	}
	return nil
}

func (*StateManager) Deduct(
	ctx context.Context,
	addr codec.Address,
	mu state.Mutable,
	amount uint64,
) error {
	_, err := BurnAsset(ctx, mu, NAIAddress, addr, amount)
	return err
}

func (*StateManager) AddBalance(
	ctx context.Context,
	addr codec.Address,
	mu state.Mutable,
	amount uint64,
	_ bool,
) error {
	_, err := MintAsset(ctx, mu, NAIAddress, addr, amount)
	return err
}

func (*StateManager) SponsorStateKeys(addr codec.Address) state.Keys {
	return state.Keys{
		string(AssetInfoKey(NAIAddress)):                 state.All,
		string(AssetAccountBalanceKey(NAIAddress, addr)): state.All,
	}
}

func (*StateManager) HeightKey() []byte {
	return []byte{heightPrefix}
}

func (*StateManager) TimestampKey() []byte {
	return []byte{timestampPrefix}
}

func (*StateManager) FeeKey() []byte {
	return []byte{feePrefix}
}
