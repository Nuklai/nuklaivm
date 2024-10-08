// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/nuklai/nuklaivm/storage"
	"github.com/nuklai/nuklaivm/vm"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/auth"
	"github.com/ava-labs/hypersdk/cli"
	"github.com/ava-labs/hypersdk/cli/prompt"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/bls"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/crypto/secp256r1"
	"github.com/ava-labs/hypersdk/utils"

	nutils "github.com/nuklai/nuklaivm/utils"
)

const (
	ed25519Key   = "ed25519"
	secp256r1Key = "secp256r1"
	blsKey       = "bls"
)

var keyCmd = &cobra.Command{
	Use: "key",
	RunE: func(*cobra.Command, []string) error {
		return ErrMissingSubcommand
	},
}

var genKeyCmd = &cobra.Command{
	Use: "generate [ed25519/secp256r1/bls]",
	PreRunE: func(_ *cobra.Command, args []string) error {
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
			"{{green}}created address:{{/}} %s\n",
			priv.Address,
		)

		// Convert the private key bytes to a base64 encoded string
		privKeyString := base64.StdEncoding.EncodeToString(priv.Bytes)
		utils.Outf("{{yellow}}Private Key String(Base64):{{/}} %s\n", privKeyString)

		// Create the directory with permissions (if it doesn't exist)
		err = os.MkdirAll("./test_accounts", 0o755)
		if err != nil {
			panic(err)
		}
		// Construct the filename with Address
		filename := fmt.Sprintf("./test_accounts/%s-%s.pk", priv.Address, args[0])
		// Write the byte slice to the file
		err = os.WriteFile(filename, priv.Bytes, 0o600)
		if err != nil {
			panic(err)
		}
		return nil
	},
}

var importKeyCmd = &cobra.Command{
	Use: "import [type] [path or encoded string]",
	PreRunE: func(_ *cobra.Command, args []string) error {
		if len(args) != 2 {
			return ErrInvalidArgs
		}
		return checkKeyType(args[0])
	},
	RunE: func(_ *cobra.Command, args []string) error {
		keyType := args[0]
		keyInput := args[1]

		var priv *cli.PrivateKey

		// Check if the provided argument is a file path or an encoded string
		if _, err := os.Stat(keyInput); err == nil {
			// It's a file path, load the private key from the file
			priv, err = loadPrivateKeyFromPath(keyType, keyInput)
			if err != nil {
				return fmt.Errorf("failed to load private key from file: %w", err)
			}
		} else {
			// It's not a valid file path, assume it's an encoded string (base64 or hex)
			priv, err = loadPrivateKeyFromString(keyType, keyInput)
			if err != nil {
				return fmt.Errorf("failed to load private key from encoded string: %w", err)
			}
		}

		// Store the private key in the key manager
		if err := handler.h.StoreKey(priv); err != nil {
			return fmt.Errorf("failed to store key: %w", err)
		}
		if err := handler.h.StoreDefaultKey(priv.Address); err != nil {
			return fmt.Errorf("failed to set default key: %w", err)
		}

		utils.Outf("{{green}}imported address:{{/}} %s\n", priv.Address)
		return nil
	},
}

var setKeyCmd = &cobra.Command{
	Use: "set",
	RunE: func(*cobra.Command, []string) error {
		return handler.SetKey()
	},
}

var balanceKeyCmd = &cobra.Command{
	Use: "balance [address]",
	RunE: func(_ *cobra.Command, args []string) error {
		var (
			addr codec.Address
			err  error
		)
		if len(args) != 1 {
			addr, _, err = handler.h.GetDefaultKey(true)
			if err != nil {
				return err
			}
		} else {
			addr, err = codec.StringToAddress(args[0])
			if err != nil {
				return err
			}
		}
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		for _, ncli := range nclients {
			if _, _, _, _, _, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(context.TODO(), ncli, addr, storage.NAIAddress, true, false, -1); err != nil {
				return err
			}
		}
		return nil
	},
}

func lookupKeyBalance(uri string, addr codec.Address, assetAddress codec.Address, isNFT bool) error {
	var err error
	if isNFT {
		_, _, _, _, _, _, _, err = handler.GetAssetNFTInfo(context.TODO(), vm.NewJSONRPCClient(uri), addr, assetAddress, true)
	} else {
		_, _, _, _, _, _, _, _, _, _, _, _, _, err = handler.GetAssetInfo(
			context.TODO(), vm.NewJSONRPCClient(uri),
			addr, assetAddress, true, false, -1)
	}
	return err
}

var balanceFTKeyCmd = &cobra.Command{
	Use: "balance-ft [address]",
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return handler.BalanceAsset(checkAllChains, false, lookupKeyBalance)
		}
		addr, err := codec.StringToAddress(args[0])
		if err != nil {
			return err
		}
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		assetAddress, err := prompt.Address("assetAddress")
		if err != nil {
			return err
		}
		for _, ncli := range nclients {
			if _, _, _, _, _, _, _, _, _, _, _, _, _, err := handler.GetAssetInfo(context.TODO(), ncli, addr, assetAddress, true, false, -1); err != nil {
				return err
			}
		}
		return nil
	},
}

var balanceNFTKeyCmd = &cobra.Command{
	Use: "balance-nft [address]",
	RunE: func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return handler.BalanceAsset(checkAllChains, true, lookupKeyBalance)
		}
		addr, err := codec.StringToAddress(args[0])
		if err != nil {
			return err
		}
		nclients, err := handler.DefaultNuklaiVMJSONRPCClient(checkAllChains)
		if err != nil {
			return err
		}
		nftAddress, err := prompt.Address("nftAddress")
		if err != nil {
			return err
		}
		for _, ncli := range nclients {
			if _, _, _, _, _, _, _, err := handler.GetAssetNFTInfo(context.TODO(), ncli, addr, nftAddress, true); err != nil {
				return err
			}
		}
		return nil
	},
}

var vanityAddressCmd = &cobra.Command{
	Use: "generate-vanity-address",
	RunE: func(_ *cobra.Command, args []string) error {
		prefix := "nuklai"
		if len(args) == 1 {
			prefix = args[0]
		}

		// Call GenerateVanityAddress to create the vanity address
		vanityAddress, err := generateVanityAddress(prefix)
		if err != nil {
			return err
		}

		// Output the generated vanity address
		utils.Outf("{{yellow}}Vanity Address: %s{{/}}\n", vanityAddress.String())

		return nil
	},
}

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
	case auth.ED25519ID:
		return ed25519Key, nil
	case auth.SECP256R1ID:
		return secp256r1Key, nil
	case auth.BLSID:
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

func loadPrivateKeyFromPath(k string, path string) (*cli.PrivateKey, error) {
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

// loadPrivateKeyFromString loads a private key from a base64 string.
func loadPrivateKeyFromString(k, keyStr string) (*cli.PrivateKey, error) {
	var decodedKey []byte
	var err error

	// Try to decode as base64 first
	decodedKey, err = base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key string: %w", err)
	}

	// Create a private key object based on the decoded bytes and key type
	switch k {
	case ed25519Key:
		pk := ed25519.PrivateKey(decodedKey)
		return &cli.PrivateKey{
			Address: auth.NewED25519Address(pk.PublicKey()),
			Bytes:   decodedKey,
		}, nil
	case secp256r1Key:
		pk := secp256r1.PrivateKey(decodedKey)
		return &cli.PrivateKey{
			Address: auth.NewSECP256R1Address(pk.PublicKey()),
			Bytes:   decodedKey,
		}, nil
	case blsKey:
		pk, err := bls.PrivateKeyFromBytes(decodedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load BLS private key: %w", err)
		}
		return &cli.PrivateKey{
			Address: auth.NewBLSAddress(bls.PublicFromPrivateKey(pk)),
			Bytes:   bls.PrivateKeyToBytes(pk),
		}, nil
	default:
		return nil, ErrInvalidKeyType
	}
}

// GenerateVanityAddress generates an address that matches the given vanity prefix.
// This address is not associated with any private key, so there is no risk of
// collision when searching for addresses.
func generateVanityAddress(prefix string) (codec.Address, error) {
	const batchSize = 1000 // Number of addresses to generate in each batch
	if len(prefix) > codec.AddressLen*2 {
		return codec.EmptyAddress, fmt.Errorf("prefix too long, max length is %d", codec.AddressLen*2)
	}

	typeID := uint8(255) // Set the typeID to 255 for a standard address
	numWorkers := runtime.NumCPU()
	fmt.Printf("Using %d workers to generate vanity address\n", numWorkers)

	resultChan := make(chan codec.Address)
	errChan := make(chan error)
	wg := sync.WaitGroup{}

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				// Generate a batch of random IDs
				for j := 0; j < batchSize; j++ {
					// Generate a random ID instead of sequentially
					randomID, err := nutils.GenerateRandomID()
					if err != nil {
						errChan <- err
						return
					}

					// Create the address using the random ID
					address := codec.CreateAddress(typeID, randomID)

					// Perform byte-level comparison for the prefix (no string conversion)
					if matchPrefix(address[:], prefix) {
						resultChan <- address
						return
					}
				}
			}
		}()
	}

	// Wait for the result or an error
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// Block until result is found
	select {
	case address := <-resultChan:
		return address, nil
	case err := <-errChan:
		return codec.EmptyAddress, err
	}
}

// matchPrefix compares the first bytes of the address against the hex representation of the prefix.
func matchPrefix(addressBytes []byte, prefix string) bool {
	// Convert prefix to bytes for comparison
	prefixBytes, _ := hex.DecodeString(prefix)

	// Compare bytes without converting the full address to a string
	if len(prefixBytes) > len(addressBytes) {
		return false
	}
	for i := 0; i < len(prefixBytes); i++ {
		if addressBytes[i] != prefixBytes[i] {
			return false
		}
	}
	return true
}
