// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/crypto/secp256r1"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/nuklai/nuklaivm/auth"
	"github.com/nuklai/nuklaivm/consts"
	brpc "github.com/nuklai/nuklaivm/rpc"
	"github.com/spf13/cobra"
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
	case consts.ED25519ID:
		return ed25519Key, nil
	case consts.SECP256R1ID:
		return secp256r1Key, nil
	case consts.BLSID:
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

func loadPrivateKey(k string, path string) (*cli.PrivateKey, error) {
	switch k {
	case ed25519Key:
		p, err := utils.LoadBytes(path, ed25519.PrivateKeyLen)
		if err != nil {
			return nil, err
		}
		pk := ed25519.PrivateKey(p)
		return &cli.PrivateKey{
			Address: auth.NewED25519Address(pk.PublicKey()),
			Bytes:   p,
		}, nil
	case secp256r1Key:
		p, err := utils.LoadBytes(path, secp256r1.PrivateKeyLen)
		if err != nil {
			return nil, err
		}
		pk := secp256r1.PrivateKey(p)
		return &cli.PrivateKey{
			Address: auth.NewSECP256R1Address(pk.PublicKey()),
			Bytes:   p,
		}, nil
	case blsKey:
		p, err := utils.LoadBytes(path, bls.PrivateKeyLen)
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
		utils.Outf(
			"{{green}}created address:{{/}} %s",
			codec.MustAddressBech32(consts.HRP, priv.Address),
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
		priv, err := loadPrivateKey(args[0], args[1])
		if err != nil {
			return err
		}
		if err := handler.h.StoreKey(priv); err != nil {
			return err
		}
		if err := handler.h.StoreDefaultKey(priv.Address); err != nil {
			return err
		}
		utils.Outf(
			"{{green}}imported address:{{/}} %s",
			codec.MustAddressBech32(consts.HRP, priv.Address),
		)
		return nil
	},
}

func lookupSetKeyBalance(choice int, address string, uri string, networkID uint32, chainID ids.ID) error {
	// TODO: just load once
	cli := brpc.NewJSONRPCClient(uri, networkID, chainID)
	balance, err := cli.Balance(context.TODO(), address)
	if err != nil {
		return err
	}
	addr, err := codec.ParseAddressBech32(consts.HRP, address)
	if err != nil {
		return err
	}
	keyType, err := getKeyType(addr)
	if err != nil {
		return err
	}
	utils.Outf(
		"%d) {{cyan}}address (%s):{{/}} %s {{cyan}}balance:{{/}} %s %s\n",
		choice,
		keyType,
		address,
		utils.FormatBalance(balance, consts.Decimals),
		consts.Symbol,
	)
	return nil
}

var setKeyCmd = &cobra.Command{
	Use: "set",
	RunE: func(*cobra.Command, []string) error {
		return handler.Root().SetKey(lookupSetKeyBalance)
	},
}

func lookupKeyBalance(addr codec.Address, uri string, networkID uint32, chainID ids.ID, _ ids.ID) error {
	_, err := handler.GetBalance(context.TODO(), brpc.NewJSONRPCClient(uri, networkID, chainID), addr)
	return err
}

var balanceKeyCmd = &cobra.Command{
	Use: "balance [address]",
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return handler.Root().Balance(checkAllChains, false, lookupKeyBalance)
		}
		addr, err := codec.ParseAddressBech32(consts.HRP, args[0])
		if err != nil {
			return err
		}
		utils.Outf("{{yellow}}address:{{/}} %s\n", args[0])
		clients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		for _, cli := range clients {
			if _, err := handler.GetBalance(context.TODO(), cli, addr); err != nil {
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
		bech32Addr, err := bech32.Encode(consts.HRP, data5Bit)
		if err != nil {
			return err
		}

		utils.Outf("{{yellow}}Address: %s{{/}}\n", bech32Addr)

		return nil
	},
}
