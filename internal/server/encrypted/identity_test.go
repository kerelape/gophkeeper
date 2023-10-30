package encrypted_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kerelape/gophkeeper/internal/server/encrypted"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/virtual"
	"github.com/stretchr/testify/assert"
)

func TestEncrypted(t *testing.T) {
	t.Run("Piece", func(t *testing.T) {
		var (
			g = encrypted.Gophkeeper{
				Origin: virtual.New(time.Hour, t.TempDir()),
				Cipher: encrypted.CFBCipher{},
			}
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)

		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		token, authenticateError := g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		identity, identityError := g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		rid, storePieceError := identity.StorePiece(
			context.Background(),
			gophkeeper.Piece{
				Meta:    "testmeta",
				Content: ([]byte)("testcontent"),
			},
			credential.Password,
		)
		assert.Nil(t, storePieceError, "expected to successfully store a piece")

		piece, restoreError := identity.RestorePiece(
			context.Background(),
			rid, credential.Password,
		)
		assert.Nil(t, restoreError, "expected to successfully restore the piece")

		assert.Equal(t, "testmeta", piece.Meta, "meta is not restored correctly")
		assert.Equal(t, "testcontent", (string)(piece.Content), "content is not restored correctly")

		origin := (identity.(encrypted.Identity)).Origin
		wrongRID, wrongRIDError := origin.StorePiece(
			context.Background(),
			gophkeeper.Piece{
				Meta:    "nonjson",
				Content: []byte{},
			},
			credential.Password,
		)
		if wrongRIDError != nil {
			t.Fail()
		}
		_, restoreWrongError := identity.RestorePiece(context.Background(), wrongRID, credential.Password)
		assert.NotNil(t, restoreWrongError)
	})

	t.Run("Blob", func(t *testing.T) {
		var (
			g = encrypted.Gophkeeper{
				Origin: virtual.New(time.Hour, t.TempDir()),
				Cipher: encrypted.CFBCipher{},
			}
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)

		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		token, authenticateError := g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		identity, identityError := g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		rid, storeBlobError := identity.StoreBlob(
			context.Background(),
			gophkeeper.Blob{
				Meta:    "testmeta",
				Content: io.NopCloser(bytes.NewReader(([]byte)("testcontent"))),
			},
			credential.Password,
		)
		assert.Nil(t, storeBlobError, "expected to successfully store a blob")

		blob, restoreError := identity.RestoreBlob(
			context.Background(),
			rid, credential.Password,
		)
		assert.Nil(t, restoreError, "expected to successfully restore the blob")

		content, contentError := io.ReadAll(blob.Content)
		assert.Nil(t, contentError, "expected to successfully read the content")

		blob.Content.Close()

		assert.Equal(t, "testmeta", blob.Meta, "meta is not restored correctly")
		assert.Equal(t, "testcontent", (string)(content), "content is not restored correctly")

		origin := (identity.(encrypted.Identity)).Origin
		wrongRID, wrongRIDError := origin.StoreBlob(
			context.Background(),
			gophkeeper.Blob{
				Meta:    "nonjson",
				Content: io.NopCloser(bytes.NewReader([]byte{})),
			},
			credential.Password,
		)
		assert.Nil(t, wrongRIDError)
		_, restoreWrongError := identity.RestoreBlob(context.Background(), wrongRID, credential.Password)
		assert.NotNil(t, restoreWrongError)
	})

	t.Run("Delete", func(t *testing.T) {
		var (
			g = encrypted.Gophkeeper{
				Origin: virtual.New(time.Hour, t.TempDir()),
				Cipher: encrypted.CFBCipher{},
			}
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)

		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		token, authenticateError := g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		identity, identityError := g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		rid, storePieceError := identity.StorePiece(
			context.Background(),
			gophkeeper.Piece{
				Meta:    "testmeta",
				Content: ([]byte)("testcontent"),
			},
			credential.Password,
		)
		assert.Nil(t, storePieceError, "expected to successfully store a piece")

		err := identity.Delete(context.Background(), rid)
		assert.Nil(t, err, "expected to successfully delete")
	})

	t.Run("List", func(t *testing.T) {
		var (
			g = encrypted.Gophkeeper{
				Origin: virtual.New(time.Hour, t.TempDir()),
				Cipher: encrypted.CFBCipher{},
			}
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)

		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		token, authenticateError := g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		identity, identityError := g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		_, storePieceError := identity.StorePiece(
			context.Background(),
			gophkeeper.Piece{
				Meta:    "testmeta",
				Content: ([]byte)("testcontent"),
			},
			credential.Password,
		)
		assert.Nil(t, storePieceError, "expected to successfully store a piece")

		resources, listError := identity.List(context.Background())
		assert.Nil(t, listError, "expected to successfully list resources")
		assert.Equal(t, 1, len(resources))
		assert.Equal(t, "testmeta", resources[0].Meta)
		assert.Equal(t, gophkeeper.ResourceTypePiece, resources[0].Type)
	})

	t.Run("Incorrect input", func(t *testing.T) {
		var (
			g = encrypted.Gophkeeper{
				Origin: virtual.New(time.Hour, t.TempDir()),
				Cipher: encrypted.CFBCipher{},
			}
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)

		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		token, authenticateError := g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		identity, identityError := g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		_, restoreIncorrectPieceError := identity.RestorePiece(context.Background(), -1, credential.Password)
		assert.NotNil(t, restoreIncorrectPieceError)

		_, restoreIncorrectBlobError := identity.RestoreBlob(context.Background(), -1, credential.Password)
		assert.NotNil(t, restoreIncorrectBlobError)

		_, invalidPasswordError := identity.StorePiece(context.Background(), gophkeeper.Piece{}, "")
		assert.NotNil(t, invalidPasswordError)

		_, invalidPasswordError = identity.StoreBlob(context.Background(), gophkeeper.Blob{Content: io.NopCloser(strings.NewReader(""))}, "")
		assert.NotNil(t, invalidPasswordError)

		_, err := (identity.(encrypted.Identity)).Origin.StorePiece(
			context.Background(),
			gophkeeper.Piece{
				Meta:    "not a json",
				Content: ([]byte)("test"),
			},
			credential.Password,
		)
		assert.NotNil(t, err)
		_, invalidContentError := identity.List(context.Background())
		assert.NotNil(t, invalidContentError)
	})
}
