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
	"strings"
	"sync"

	"github.com/nuklai/nuklaivm/vm"
	"github.com/spf13/cobra"

	"github.com/ava-labs/hypersdk/auth"
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

		var priv *auth.PrivateKey

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
		ctx := context.Background()
		_, _, _, _, bcli, _, err := handler.DefaultActor()
		if err != nil {
			return err
		}
		var addr codec.Address
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
		_, err = handler.GetBalance(ctx, bcli, addr)
		if err != nil {
			return err
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

var balanceAssetKeyCmd = &cobra.Command{
	Use: "balance-asset [address]",
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
	Use: "nft [address]",
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

// generatePrivateKey generates a private key and displays it in base64 and hex.
func generatePrivateKey(k string) (*auth.PrivateKey, error) {
	var priv *auth.PrivateKey
	var keyBytes []byte

	switch k {
	case ed25519Key:
		p, err := ed25519.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}

		// Use the full 64 bytes for Ed25519
		keyBytes = p[:ed25519.PrivateKeyLen] // 64 bytes
		priv = &auth.PrivateKey{
			Address: auth.NewED25519Address(p.PublicKey()),
			Bytes:   keyBytes,
		}
	case secp256r1Key:
		p, err := secp256r1.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		// Use only 32 bytes for secp256r1
		keyBytes = p[:secp256r1.PrivateKeyLen] // 32 bytes
		priv = &auth.PrivateKey{
			Address: auth.NewSECP256R1Address(p.PublicKey()),
			Bytes:   keyBytes,
		}
	case blsKey:
		p, err := bls.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		// Ensure exactly 32 bytes for BLS
		keyBytes = bls.PrivateKeyToBytes(p)[:bls.PrivateKeyLen] // 32 bytes
		priv = &auth.PrivateKey{
			Address: auth.NewBLSAddress(bls.PublicFromPrivateKey(p)),
			Bytes:   keyBytes,
		}
	default:
		return nil, ErrInvalidKeyType
	}

	// Display private key in base64 and hex formats
	privKeyBase64 := base64.StdEncoding.EncodeToString(priv.Bytes)
	utils.Outf("{{yellow}}Private Key (Base64):{{/}} %s\n", privKeyBase64)

	privKeyHex := codec.ToHex(priv.Bytes)
	utils.Outf("{{yellow}}Private Key (Hex):{{/}} %s\n", privKeyHex)

	// Optionally save private key to a file if needed
	filename := fmt.Sprintf("./test_accounts/%s-%s.pk", priv.Address, k)
	err := os.WriteFile(filename, priv.Bytes, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to write private key to file: %w", err)
	}

	return priv, nil
}

// loadPrivateKeyFromString loads a private key from a base64 or hex string,
// verifying it matches the expected length for the specified key type.
func loadPrivateKeyFromString(k, keyStr string) (*auth.PrivateKey, error) {
	var decodedKey []byte
	var err error
	var privKeyLength int

	switch k {
	case ed25519Key:
		privKeyLength = ed25519.PrivateKeyLen
	case secp256r1Key:
		privKeyLength = secp256r1.PrivateKeyLen
	case blsKey:
		privKeyLength = bls.PrivateKeyLen
	default:
		return nil, ErrInvalidKeyType
	}

	// Attempt to decode as base64 first
	decodedKey, err = base64.StdEncoding.DecodeString(keyStr)
	if err == nil && len(decodedKey) == privKeyLength {
		// Successfully decoded as base64 and the length is correct
		fmt.Println("Decoded key as base64 successfully.")
	} else {
		// If base64 decoding fails, try hex decoding
		decodedKey, err = codec.LoadHex(strings.TrimSpace(keyStr), privKeyLength)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key string: input is not valid hex or length mismatch")
		}
		fmt.Println("Decoded key as hex successfully.")
	}

	// Ensure key length matches expected length for each type
	switch k {
	case ed25519Key:
		pk := ed25519.PrivateKey(decodedKey)
		return &auth.PrivateKey{
			Address: auth.NewED25519Address(pk.PublicKey()),
			Bytes:   decodedKey,
		}, nil
	case secp256r1Key:
		pk := secp256r1.PrivateKey(decodedKey)
		return &auth.PrivateKey{
			Address: auth.NewSECP256R1Address(pk.PublicKey()),
			Bytes:   decodedKey,
		}, nil
	case blsKey:
		pk, err := bls.PrivateKeyFromBytes(decodedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load BLS private key: %w", err)
		}
		return &auth.PrivateKey{
			Address: auth.NewBLSAddress(bls.PublicFromPrivateKey(pk)),
			Bytes:   bls.PrivateKeyToBytes(pk),
		}, nil
	default:
		return nil, ErrInvalidKeyType
	}
}

func loadPrivateKeyFromPath(k string, path string) (*auth.PrivateKey, error) {
	switch k {
	case ed25519Key:
		p, err := utils.LoadBytes(path, ed25519.PrivateKeyLen)
		if err != nil {
			return nil, err
		}
		pk := ed25519.PrivateKey(p)
		return &auth.PrivateKey{
			Address: auth.NewED25519Address(pk.PublicKey()),
			Bytes:   p,
		}, nil
	case secp256r1Key:
		p, err := utils.LoadBytes(path, secp256r1.PrivateKeyLen)
		if err != nil {
			return nil, err
		}
		pk := secp256r1.PrivateKey(p)
		return &auth.PrivateKey{
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
		return &auth.PrivateKey{
			Address: auth.NewBLSAddress(bls.PublicFromPrivateKey(privKey)),
			Bytes:   p,
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
