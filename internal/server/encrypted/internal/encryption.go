// Package encryption provides encryption utils.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/pbkdf2"
)

// Data is encryption data.
type Data struct {
	Block cipher.Block
	IV    []byte
	Salt  []byte
}

// Password returns password based encryption data.
func Password(password string) (Data, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return Data{}, err
	}

	block, blockError := aes.NewCipher(
		pbkdf2.Key(([]byte)(password), salt, 4096, 32, sha256.New),
	)
	if blockError != nil {
		return Data{}, blockError
	}

	iv := make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return Data{}, err
	}

	return Data{block, iv, salt}, nil
}
