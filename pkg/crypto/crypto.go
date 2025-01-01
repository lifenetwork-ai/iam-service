package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/go-bip39"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
)

func GenerateAccount(mnemonic, passphrase, salt, accountType string, id uint64) (*accounts.Account, *ecdsa.PrivateKey, error) {
	// Generate the seed from the mnemonic and passphrase
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Create the master key from the seed
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, nil, err
	}

	// Convert walletType to a unique integer for use in the HD Path
	accountTypeHash := hashToUint32(accountType + fmt.Sprint(id))

	// Define the HD Path for Ethereum address (e.g., m/44'/60'/id'/walletTypeHash/salt)
	path := []uint32{
		44 + bip32.FirstHardenedChild,         // BIP44 purpose field
		60 + bip32.FirstHardenedChild,         // Ethereum coin type
		uint32(id) + bip32.FirstHardenedChild, // User-specific field
		accountTypeHash,                       // Unique integer based on account type and id
		hashToUint32(salt),                    // Hash of salt for additional security
	}

	// Derive a private key along the specified HD Path
	key := masterKey
	for _, index := range path {
		key, err = key.NewChildKey(index)
		if err != nil {
			return nil, nil, err
		}
	}

	// Generate an Ethereum account from the derived private key
	privateKey, err := crypto.ToECDSA(key.Key)
	if err != nil {
		return nil, nil, err
	}

	account := accounts.Account{
		Address: crypto.PubkeyToAddress(privateKey.PublicKey),
	}

	return &account, privateKey, nil
}

// hashToUint32 generates a unique integer from a string input
func hashToUint32(input string) uint32 {
	hash := sha256.Sum256([]byte(input))
	return binary.BigEndian.Uint32(hash[:4])
}
