package encrypted

import "crypto/cipher"

// CFBCipher is a CFB Cipher.
type CFBCipher struct{}

var _ Cipher = (*CFBCipher)(nil)

// Encrypter implements Cipher.
func (c CFBCipher) Encrypter(block cipher.Block, iv []byte) cipher.Stream {
	return cipher.NewCFBEncrypter(block, iv)
}

// Decrypter implements Cipher.
func (c CFBCipher) Decrypter(block cipher.Block, iv []byte) cipher.Stream {
	return cipher.NewCFBDecrypter(block, iv)
}
