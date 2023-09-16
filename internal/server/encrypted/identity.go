package encrypted

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"

	composedreadcloser "github.com/kerelape/gophkeeper/internal/composed_read_closer"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"golang.org/x/crypto/pbkdf2"
)

const (
	keyLen  = 32
	keyIter = 4096
)

type meta struct {
	IV      []byte `json:"iv"`
	Salt    []byte `json:"salt"`
	Content string `json:"content"`
}

// Cipher is a factory of encrypter and pairing decrypters.
type Cipher interface {
	// Encrypter returns encrypting Stream.
	Encrypter(block cipher.Block, iv []byte) cipher.Stream

	// Decrypter returns decrypting Stream.
	Decrypter(block cipher.Block, iv []byte) cipher.Stream
}

// Identity is an encrpypted gophkeeper Identity.
type Identity struct {
	Origin gophkeeper.Identity
	Cipher Cipher
}

var _ gophkeeper.Identity = (*Identity)(nil)

// StorePiece implements gophkeeper.Identity.
func (i Identity) StorePiece(ctx context.Context, piece gophkeeper.Piece, password string) (gophkeeper.ResourceID, error) {
	var (
		salt []byte = make([]byte, 8)
		iv   []byte = make([]byte, 12)
	)
	if _, err := rand.Read(salt); err != nil {
		return -1, err
	}
	if _, err := rand.Read(iv); err != nil {
		return -1, err
	}

	var block, blockError = aes.NewCipher(
		pbkdf2.Key(([]byte)(password), salt, keyIter, keyLen, sha256.New),
	)
	if blockError != nil {
		return -1, blockError
	}

	var reader = cipher.StreamReader{
		S: i.Cipher.Encrypter(block, iv),
		R: bytes.NewReader(piece.Content),
	}
	var content, contentError = io.ReadAll(reader)
	if contentError != nil {
		return -1, contentError
	}
	piece.Content = content

	var meta, metaError = json.Marshal(
		meta{
			IV:      iv,
			Salt:    salt,
			Content: piece.Meta,
		},
	)
	if metaError != nil {
		return -1, metaError
	}
	piece.Meta = (string)(meta)

	return i.Origin.StorePiece(ctx, piece, password)
}

// RestorePiece implements gophkeeper.Identity.
func (i Identity) RestorePiece(ctx context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Piece, error) {
	var piece, pieceError = i.Origin.RestorePiece(ctx, rid, password)
	if pieceError != nil {
		return gophkeeper.Piece{}, pieceError
	}

	var meta meta
	if err := json.Unmarshal(([]byte)(piece.Meta), &meta); err != nil {
		return gophkeeper.Piece{}, err
	}
	piece.Meta = meta.Content

	var block, blockError = aes.NewCipher(
		pbkdf2.Key(([]byte)(password), meta.Salt, keyIter, keyLen, sha256.New),
	)
	if blockError != nil {
		return gophkeeper.Piece{}, blockError
	}

	var reader = cipher.StreamReader{
		S: i.Cipher.Decrypter(block, meta.IV),
		R: bytes.NewReader(piece.Content),
	}
	var content, contentError = io.ReadAll(reader)
	if contentError != nil {
		return gophkeeper.Piece{}, contentError
	}
	piece.Content = content

	return piece, nil
}

// StoreBlob implements gophkeeper.Identity.
func (i Identity) StoreBlob(ctx context.Context, blob gophkeeper.Blob, password string) (gophkeeper.ResourceID, error) {
	var salt []byte = make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return -1, err
	}
	var block, blockError = aes.NewCipher(
		pbkdf2.Key(([]byte)(password), salt, keyIter, keyLen, sha256.New),
	)
	if blockError != nil {
		return -1, blockError
	}
	var iv []byte = make([]byte, block.BlockSize())
	if _, err := rand.Read(iv); err != nil {
		return -1, err
	}

	var meta, metaError = json.Marshal(
		meta{
			IV:      iv,
			Salt:    salt,
			Content: blob.Meta,
		},
	)
	if metaError != nil {
		return -1, metaError
	}

	return i.Origin.StoreBlob(
		ctx,
		gophkeeper.Blob{
			Meta: (string)(meta),
			Content: &composedreadcloser.ComposedReadCloser{
				Reader: cipher.StreamReader{
					S: i.Cipher.Encrypter(
						block,
						iv,
					),
					R: blob.Content,
				},
				Closer: blob.Content,
			},
		},
		password,
	)
}

// RestoreBlob implements gophkeeper.Identity.
func (i Identity) RestoreBlob(ctx context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Blob, error) {
	var blob, blobError = i.Origin.RestoreBlob(ctx, rid, password)
	if blobError != nil {
		return gophkeeper.Blob{}, blobError
	}

	var meta meta
	if err := json.Unmarshal(([]byte)(blob.Meta), &meta); err != nil {
		return gophkeeper.Blob{}, err
	}

	var block, blockError = aes.NewCipher(
		pbkdf2.Key(([]byte)(password), meta.Salt, 0, 0, sha256.New),
	)
	if blockError != nil {
		return gophkeeper.Blob{}, blockError
	}

	var decryptedBlob = gophkeeper.Blob{
		Meta: meta.Content,
		Content: &composedreadcloser.ComposedReadCloser{
			Reader: cipher.StreamReader{
				S: i.Cipher.Decrypter(block, meta.IV),
				R: blob.Content,
			},
			Closer: blob.Content,
		},
	}
	return decryptedBlob, nil
}

// List implements gophkeeper.Identity.
func (i Identity) List(ctx context.Context) ([]gophkeeper.Resource, error) {
	var resources, resourcesError = i.Origin.List(ctx)
	if resourcesError != nil {
		return nil, resourcesError
	}
	for i := range resources {
		var (
			resource = &resources[i]
			meta     meta
		)
		if err := json.Unmarshal(([]byte)(resource.Meta), &meta); err != nil {
			return nil, err
		}
		resource.Meta = meta.Content
	}
	return resources, nil
}

// Delete implements gophkeeper.Identity.
func (i Identity) Delete(ctx context.Context, rid gophkeeper.ResourceID) error {
	return i.Origin.Delete(ctx, rid)
}
