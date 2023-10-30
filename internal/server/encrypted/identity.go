package encrypted

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/json"
	"io"

	composedreadcloser "github.com/kerelape/gophkeeper/internal/composed_read_closer"
	encryption "github.com/kerelape/gophkeeper/internal/server/encrypted/internal"
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
	enc, encError := encryption.Password(password)
	if encError != nil {
		return -1, encError
	}

	reader := cipher.StreamReader{
		S: i.Cipher.Encrypter(enc.Block, enc.IV),
		R: bytes.NewReader(piece.Content),
	}
	content, contentError := io.ReadAll(reader)
	if contentError != nil {
		return -1, contentError
	}
	piece.Content = content

	wrappedMeta, wrappedMetaError := json.Marshal(
		meta{
			IV:      enc.IV,
			Salt:    enc.Salt,
			Content: piece.Meta,
		},
	)
	if wrappedMetaError != nil {
		return -1, wrappedMetaError
	}
	piece.Meta = (string)(wrappedMeta)

	return i.Origin.StorePiece(ctx, piece, password)
}

// RestorePiece implements gophkeeper.Identity.
func (i Identity) RestorePiece(ctx context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Piece, error) {
	piece, pieceError := i.Origin.RestorePiece(ctx, rid, password)
	if pieceError != nil {
		return gophkeeper.Piece{}, pieceError
	}

	var m meta
	if err := json.Unmarshal(([]byte)(piece.Meta), &m); err != nil {
		return gophkeeper.Piece{}, err
	}
	piece.Meta = m.Content

	block, blockError := aes.NewCipher(
		pbkdf2.Key(([]byte)(password), m.Salt, keyIter, keyLen, sha256.New),
	)
	if blockError != nil {
		return gophkeeper.Piece{}, blockError
	}

	reader := cipher.StreamReader{
		S: i.Cipher.Decrypter(block, m.IV),
		R: bytes.NewReader(piece.Content),
	}
	content, contentError := io.ReadAll(reader)
	if contentError != nil {
		return gophkeeper.Piece{}, contentError
	}
	piece.Content = content

	return piece, nil
}

// StoreBlob implements gophkeeper.Identity.
func (i Identity) StoreBlob(ctx context.Context, blob gophkeeper.Blob, password string) (gophkeeper.ResourceID, error) {
	enc, encError := encryption.Password(password)
	if encError != nil {
		return -1, encError
	}

	m, metaError := json.Marshal(
		meta{
			IV:      enc.IV,
			Salt:    enc.Salt,
			Content: blob.Meta,
		},
	)
	if metaError != nil {
		return -1, metaError
	}

	return i.Origin.StoreBlob(
		ctx,
		gophkeeper.Blob{
			Meta: (string)(m),
			Content: &composedreadcloser.ComposedReadCloser{
				Reader: cipher.StreamReader{
					S: i.Cipher.Encrypter(enc.Block, enc.IV),
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
	blob, blobError := i.Origin.RestoreBlob(ctx, rid, password)
	if blobError != nil {
		return gophkeeper.Blob{}, blobError
	}

	var m meta
	if err := json.Unmarshal(([]byte)(blob.Meta), &m); err != nil {
		return gophkeeper.Blob{}, err
	}

	block, blockError := aes.NewCipher(
		pbkdf2.Key(([]byte)(password), m.Salt, keyIter, keyLen, sha256.New),
	)
	if blockError != nil {
		return gophkeeper.Blob{}, blockError
	}

	decryptedBlob := gophkeeper.Blob{
		Meta: m.Content,
		Content: &composedreadcloser.ComposedReadCloser{
			Reader: cipher.StreamReader{
				S: i.Cipher.Decrypter(block, m.IV),
				R: blob.Content,
			},
			Closer: blob.Content,
		},
	}
	return decryptedBlob, nil
}

// List implements gophkeeper.Identity.
func (i Identity) List(ctx context.Context) ([]gophkeeper.Resource, error) {
	resources, resourcesError := i.Origin.List(ctx)
	if resourcesError != nil {
		return nil, resourcesError
	}
	for i := range resources {
		var (
			resource = &resources[i]
			m        meta
		)
		if err := json.Unmarshal(([]byte)(resource.Meta), &m); err != nil {
			return nil, err
		}
		resource.Meta = m.Content
	}
	return resources, nil
}

// Delete implements gophkeeper.Identity.
func (i Identity) Delete(ctx context.Context, rid gophkeeper.ResourceID) error {
	return i.Origin.Delete(ctx, rid)
}
