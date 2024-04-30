// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/crypto/secp256r1"
	hutils "github.com/ava-labs/hypersdk/utils"
	"github.com/btcsuite/btcd/btcutil/bech32"

	"github.com/nuklai/nuklaivm/auth"
	"github.com/nuklai/nuklaivm/challenge"
	frpc "github.com/nuklai/nuklaivm/cmd/nuklai-faucet/rpc"
	nconsts "github.com/nuklai/nuklaivm/consts"
	nrpc "github.com/nuklai/nuklaivm/rpc"
)

const (
	ed25519Key   = "ed25519"
	secp256r1Key = "secp256r1"
	blsKey       = "bls"
)

func checkKeyType(k string) error {
	switch k {
	case ed25519Key, secp256r1Key, blsKey:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidKeyType, k)
	}
}

func getKeyType(addr codec.Address) (string, error) {
	switch addr[0] {
	case nconsts.ED25519ID:
		return ed25519Key, nil
	case nconsts.SECP256R1ID:
		return secp256r1Key, nil
	case nconsts.BLSID:
		return blsKey, nil
	default:
		return "", ErrInvalidKeyType
	}
}

func generatePrivateKey(k string) (*cli.PrivateKey, error) {
	switch k {
	case ed25519Key:
		p, err := ed25519.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		return &cli.PrivateKey{
			Address: auth.NewED25519Address(p.PublicKey()),
			Bytes:   p[:],
		}, nil
	case secp256r1Key:
		p, err := secp256r1.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		return &cli.PrivateKey{
			Address: auth.NewSECP256R1Address(p.PublicKey()),
			Bytes:   p[:],
		}, nil
	case blsKey:
		p, err := bls.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		return &cli.PrivateKey{
			Address: auth.NewBLSAddress(bls.PublicFromPrivateKey(p)),
			Bytes:   bls.PrivateKeyToBytes(p),
		}, nil
	default:
		return nil, ErrInvalidKeyType
	}
}

func LoadPrivateKey(k string, path string) (*cli.PrivateKey, error) {
	switch k {
	case ed25519Key:
		p, err := hutils.LoadBytes(path, ed25519.PrivateKeyLen)
		if err != nil {
			return nil, err
		}
		pk := ed25519.PrivateKey(p)
		return &cli.PrivateKey{
			Address: auth.NewED25519Address(pk.PublicKey()),
			Bytes:   p,
		}, nil
	case secp256r1Key:
		p, err := hutils.LoadBytes(path, secp256r1.PrivateKeyLen)
		if err != nil {
			return nil, err
		}
		pk := secp256r1.PrivateKey(p)
		return &cli.PrivateKey{
			Address: auth.NewSECP256R1Address(pk.PublicKey()),
			Bytes:   p,
		}, nil
	case blsKey:
		p, err := hutils.LoadBytes(path, bls.PrivateKeyLen)
		if err != nil {
			return nil, err
		}

		privKey, err := bls.PrivateKeyFromBytes(p)
		if err != nil {
			return nil, err
		}
		return &cli.PrivateKey{
			Address: auth.NewBLSAddress(bls.PublicFromPrivateKey(privKey)),
			Bytes:   p,
		}, nil
	default:
		return nil, ErrInvalidKeyType
	}
}

var keyCmd = &cobra.Command{
	Use: "key",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var genKeyCmd = &cobra.Command{
	Use: "generate [ed25519/secp256r1/bls]",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return ErrInvalidArgs
		}
		return checkKeyType(args[0])
	},
	RunE: func(_ *cobra.Command, args []string) error {
		priv, err := generatePrivateKey(args[0])
		if err != nil {
			return err
		}
		if err := handler.h.StoreKey(priv); err != nil {
			return err
		}
		if err := handler.h.StoreDefaultKey(priv.Address); err != nil {
			return err
		}
		hutils.Outf(
			"{{green}}created address:{{/}} %s",
			codec.MustAddressBech32(nconsts.HRP, priv.Address),
		)
		return nil
	},
}

var importKeyCmd = &cobra.Command{
	Use: "import [type] [path]",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return ErrInvalidArgs
		}
		return checkKeyType(args[0])
	},
	RunE: func(_ *cobra.Command, args []string) error {
		priv, err := LoadPrivateKey(args[0], args[1])
		if err != nil {
			return err
		}
		if err := handler.h.StoreKey(priv); err != nil {
			return err
		}
		if err := handler.h.StoreDefaultKey(priv.Address); err != nil {
			return err
		}
		hutils.Outf(
			"{{green}}imported address:{{/}} %s",
			codec.MustAddressBech32(nconsts.HRP, priv.Address),
		)
		return nil
	},
}

func lookupSetKeyBalance(choice int, address string, uri string, networkID uint32, chainID ids.ID) error {
	// TODO: just load once
	ncli := nrpc.NewJSONRPCClient(uri, networkID, chainID)
	balance, err := ncli.Balance(context.TODO(), address, ids.Empty)
	if err != nil {
		return err
	}
	addr, err := codec.ParseAddressBech32(nconsts.HRP, address)
	if err != nil {
		return err
	}
	keyType, err := getKeyType(addr)
	if err != nil {
		return err
	}
	hutils.Outf(
		"%d) {{cyan}}address (%s):{{/}} %s {{cyan}}balance:{{/}} %s %s\n",
		choice,
		keyType,
		address,
		hutils.FormatBalance(balance, nconsts.Decimals),
		nconsts.Symbol,
	)
	return nil
}

var setKeyCmd = &cobra.Command{
	Use: "set",
	RunE: func(*cobra.Command, []string) error {
		return handler.Root().SetKey(lookupSetKeyBalance)
	},
}

func lookupKeyBalance(addr codec.Address, uri string, networkID uint32, chainID ids.ID, assetID ids.ID) error {
	_, _, _, _, err := handler.GetAssetInfo(
		context.TODO(), nrpc.NewJSONRPCClient(uri, networkID, chainID),
		addr, assetID, true)
	return err
}

var balanceKeyCmd = &cobra.Command{
	Use: "balance [address]",
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return handler.Root().Balance(checkAllChains, true, lookupKeyBalance)
		}
		addr, err := codec.ParseAddressBech32(nconsts.HRP, args[0])
		if err != nil {
			return err
		}
		hutils.Outf("{{yellow}}address:{{/}} %s\n", args[0])
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		assetID, err := handler.h.PromptAsset("assetID", true)
		if err != nil {
			return err
		}
		for _, ncli := range nclients {
			if _, _, _, _, err := handler.GetAssetInfo(context.TODO(), ncli, addr, assetID, true); err != nil {
				return err
			}
		}
		return nil
	},
}

func generateRandomData(n int) ([]byte, error) {
	data := make([]byte, n)
	_, err := rand.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var vanityAddressCmd = &cobra.Command{
	Use: "generate-vanity-address",
	RunE: func(_ *cobra.Command, args []string) error {
		randomData, err := generateRandomData(19) // Generate 19 random bytes
		if err != nil {
			return err
		}

		// Define a clear special pattern for the data part
		dataPart := append([]byte("nuklaivmvanity"), randomData...)

		// Convert data to 5-bit words as required by Bech32
		data5Bit, err := bech32.ConvertBits(dataPart, 8, 5, true)
		if err != nil {
			return err
		}

		// Encode to Bech32
		bech32Addr, err := bech32.Encode(nconsts.HRP, data5Bit)
		if err != nil {
			return err
		}

		hutils.Outf("{{yellow}}Address: %s{{/}}\n", bech32Addr)

		return nil
	},
}

var faucetKeyCmd = &cobra.Command{
	Use: "faucet",
	RunE: func(*cobra.Command, []string) error {
		ctx := context.Background()

		// Get private key
		_, priv, _, _, _, _, err := handler.DefaultActor()
		if err != nil {
			return err
		}

		// Get faucet
		faucetURI, err := handler.Root().PromptString("faucet URI", 0, consts.MaxInt)
		if err != nil {
			return err
		}
		fcli := frpc.NewJSONRPCClient(faucetURI)
		faucet, err := fcli.FaucetAddress(ctx)
		if err != nil {
			return err
		}

		// Search for funds
		salt, difficulty, err := fcli.Challenge(ctx)
		if err != nil {
			return err
		}
		hutils.Outf("{{yellow}}searching for faucet solutions (difficulty=%d, faucet=%s):{{/}} %x\n", difficulty, faucet, salt)
		start := time.Now()
		solution, attempts := challenge.Search(salt, difficulty, numCores)
		hutils.Outf("{{cyan}}found solution (attempts=%d, t=%s):{{/}} %x\n", attempts, time.Since(start), solution)
		txID, amount, err := fcli.SolveChallenge(ctx, codec.MustAddressBech32(nconsts.HRP, priv.Address), salt, solution)
		if err != nil {
			return err
		}
		hutils.Outf("{{green}}faucet funds incoming (%s %s):{{/}} %s\n", hutils.FormatBalance(amount, nconsts.Decimals), nconsts.Symbol, txID)
		return nil
	},
}
