package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/cosmos/go-bip39"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
)

func GenerateAccount(mnemonic, passphrase, salt, accountType, id string) (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	// Generate the seed from the mnemonic and passphrase
	seed := bip39.NewSeed(mnemonic, passphrase)

	// Create the master key from the seed
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, nil, err
	}

	// Convert walletType to a unique integer for use in the HD Path
	accountTypeHash := hashToUint32(accountType + id)

	// Define the HD Path for Ethereum address (e.g., m/44'/60'/id'/walletTypeHash/salt)
	path := []uint32{
		44 + bip32.FirstHardenedChild,               // BIP44 purpose field
		60 + bip32.FirstHardenedChild,               // Ethereum coin type
		hashToUint32(id) + bip32.FirstHardenedChild, // User-specific field
		accountTypeHash,                             // Unique integer based on account type and id
		hashToUint32(salt),                          // Hash of salt for additional security
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

	return &privateKey.PublicKey, privateKey, nil
}

// hashToUint32 generates a unique integer from a string input
func hashToUint32(input string) uint32 {
	hash := sha256.Sum256([]byte(input))
	return binary.BigEndian.Uint32(hash[:4])
}

// PublicKeyToHex converts an ECDSA public key to a hexadecimal string.
func PublicKeyToHex(publicKey *ecdsa.PublicKey) (string, error) {
	if publicKey == nil {
		return "", errors.New("public key is nil")
	}

	// Uncompressed format: 0x04 || X || Y
	// 0x04 is the prefix for uncompressed keys
	xBytes := publicKey.X.Bytes()
	yBytes := publicKey.Y.Bytes()

	// Ensure fixed-length encoding (pad to 32 bytes if needed)
	xBytesPadded := make([]byte, 32)
	yBytesPadded := make([]byte, 32)
	copy(xBytesPadded[32-len(xBytes):], xBytes)
	copy(yBytesPadded[32-len(yBytes):], yBytes)

	// Combine the prefix (0x04), X, and Y
	uncompressed := append([]byte{0x04}, append(xBytesPadded, yBytesPadded...)...)

	// Convert to hex string
	return hex.EncodeToString(uncompressed), nil
}
