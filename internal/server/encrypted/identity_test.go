package encrypted_test

import (
	"bytes"
	"context"
	"io"
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

		var registerError = g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		var token, authenticateError = g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		var identity, identityError = g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		var rid, storePieceError = identity.StorePiece(
			context.Background(),
			gophkeeper.Piece{
				Meta:    "testmeta",
				Content: ([]byte)("testcontent"),
			},
			credential.Password,
		)
		assert.Nil(t, storePieceError, "expected to successfully store a piece")

		var piece, restoreError = identity.RestorePiece(
			context.Background(),
			rid, credential.Password,
		)
		assert.Nil(t, restoreError, "expected to successfully restore the piece")

		assert.Equal(t, "testmeta", piece.Meta, "meta is not restored correctly")
		assert.Equal(t, "testcontent", (string)(piece.Content), "content is not restored correctly")
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

		var registerError = g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		var token, authenticateError = g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		var identity, identityError = g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		var rid, storeBlobError = identity.StoreBlob(
			context.Background(),
			gophkeeper.Blob{
				Meta:    "testmeta",
				Content: io.NopCloser(bytes.NewReader(([]byte)("testcontent"))),
			},
			credential.Password,
		)
		assert.Nil(t, storeBlobError, "expected to successfully store a blob")

		var blob, restoreError = identity.RestoreBlob(
			context.Background(),
			rid, credential.Password,
		)
		assert.Nil(t, restoreError, "expected to successfully restore the blob")

		var content, contentError = io.ReadAll(blob.Content)
		assert.Nil(t, contentError, "expected to successfully read the content")

		blob.Content.Close()

		assert.Equal(t, "testmeta", blob.Meta, "meta is not restored correctly")
		assert.Equal(t, "testcontent", (string)(content), "content is not restored correctly")
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

		var registerError = g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		var token, authenticateError = g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		var identity, identityError = g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		var rid, storePieceError = identity.StorePiece(
			context.Background(),
			gophkeeper.Piece{
				Meta:    "testmeta",
				Content: ([]byte)("testcontent"),
			},
			credential.Password,
		)
		assert.Nil(t, storePieceError, "expected to successfully store a piece")

		var err = identity.Delete(context.Background(), rid)
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

		var registerError = g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		var token, authenticateError = g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authenticate")

		var identity, identityError = g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get the identity")

		var _, storePieceError = identity.StorePiece(
			context.Background(),
			gophkeeper.Piece{
				Meta:    "testmeta",
				Content: ([]byte)("testcontent"),
			},
			credential.Password,
		)
		assert.Nil(t, storePieceError, "expected to successfully store a piece")

		var resources, listError = identity.List(context.Background())
		assert.Nil(t, listError, "expected to successfully list resources")
		assert.Equal(t, 1, len(resources))
		assert.Equal(t, "testmeta", resources[0].Meta)
		assert.Equal(t, gophkeeper.ResourceTypePiece, resources[0].Type)
	})
}
