package virtual_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/virtual"
	"github.com/stretchr/testify/assert"
)

func TestIdentity(t *testing.T) {
	var (
		g          = virtual.New(time.Hour, t.TempDir())
		credential = gophkeeper.Credential{
			Username: "test",
			Password: "qwerty",
		}
		alianCredential = gophkeeper.Credential{
			Username: "abc",
			Password: "nonword",
		}
	)
	registerError := g.Register(context.Background(), credential)
	assert.Nil(t, registerError, "expected to successfully register")

	registerAlianError := g.Register(context.Background(), alianCredential)
	assert.Nil(t, registerAlianError, "expected to successfully register")

	token, authenticateError := g.Authenticate(context.Background(), credential)
	assert.Nil(t, authenticateError, "expected to successfully authenticate")

	alianToken, authenticateAlianError := g.Authenticate(context.Background(), alianCredential)
	assert.Nil(t, authenticateAlianError, "expected to successfully authenticate")

	identity, identityError := g.Identity(context.Background(), token)
	assert.Nil(t, identityError, "expected to successfully get the identity")

	alian, alianError := g.Identity(context.Background(), alianToken)
	assert.Nil(t, alianError, "expected to successfully get the identity")

	rid, storeError := identity.StorePiece(
		context.Background(),
		gophkeeper.Piece{
			Meta:    "testmeta",
			Content: ([]byte)("testcontent"),
		},
		credential.Password,
	)
	assert.Nil(t, storeError, "expected to successfully store a piece")
	piece, restoreError := identity.RestorePiece(
		context.Background(),
		rid, credential.Password,
	)
	assert.Nil(t, restoreError, "expected to successfully restore the piece back")
	assert.Equal(t, "testmeta", piece.Meta, "meta is not restored correctly")
	assert.Equal(t, "testcontent", (string)(piece.Content), "content is not restored correctly")

	_, alianRestoreError := alian.RestorePiece(
		context.Background(),
		rid, alianCredential.Password,
	)
	assert.NotNil(t, alianRestoreError, "expected to get an error on restoring other identity's piece")
	assert.ErrorIs(
		t,
		alianRestoreError,
		gophkeeper.ErrResourceNotFound,
		"unexpected error on restoring other identity's piece",
	)

	blobRID, storeBlobError := identity.StoreBlob(
		context.Background(),
		gophkeeper.Blob{
			Meta:    "testmeta",
			Content: io.NopCloser(bytes.NewReader(([]byte)("testcontent"))),
		},
		credential.Password,
	)
	assert.Nil(t, storeBlobError, "expected to successfully store a blob")

	blob, restoreBlobError := identity.RestoreBlob(
		context.Background(),
		blobRID, credential.Password,
	)
	assert.Nil(t, restoreBlobError, "expected to successfully restore a blob")
	assert.Equal(t, "testmeta", blob.Meta, "meta is not restored correctly")
	assert.Equal(
		t,
		"testcontent",
		func() string {
			content, err := io.ReadAll(blob.Content)
			assert.Nil(t, err, "expected to successfully read content")
			return (string)(content)
		}(),
		"content is not restored correctly",
	)
	assert.Nil(t, blob.Content.Close(), "failed to close blob content")

	_, restoreAlianBlob := alian.RestoreBlob(
		context.Background(),
		blobRID, alianCredential.Password,
	)
	assert.NotNil(t, restoreAlianBlob, "expected to get an error on restoring alian blob")
	assert.ErrorIs(t, restoreAlianBlob, gophkeeper.ErrResourceNotFound, "unexpected error")

	deleteError := identity.Delete(context.Background(), rid)
	assert.Nil(t, deleteError, "did not expect an error")

	alianDeleteError := alian.Delete(context.Background(), blobRID)
	assert.NotNil(t, alianDeleteError, "expected an error")
	assert.ErrorIs(t, alianDeleteError, gophkeeper.ErrResourceNotFound, "unexpected error")

	resources, listError := identity.List(context.Background())
	assert.Nil(t, listError, "did not expect an error")
	assert.Equal(t, 1, len(resources), "expected resources list to be equal to 2")

	t.Run("Invalid RID", func(t *testing.T) {
		const RID = -1
		t.Run("Restore piece", func(t *testing.T) {
			_, err := identity.RestorePiece(context.Background(), RID, credential.Password)
			assert.NotNil(t, err, "expected to get an error")
			assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
		})
		t.Run("Restore blob", func(t *testing.T) {
			_, err := identity.RestoreBlob(context.Background(), RID, credential.Password)
			assert.NotNil(t, err, "expected to get an error")
			assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
		})
		t.Run("Delete", func(t *testing.T) {
			err := identity.Delete(context.Background(), RID)
			assert.NotNil(t, err, "expected to get an error")
			assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
		})
	})

	t.Run("Bad password", func(t *testing.T) {
		_, storePieceError := identity.StorePiece(context.Background(), gophkeeper.Piece{}, "")
		assert.NotNil(t, storePieceError)

		_, storeBlobError := identity.StoreBlob(context.Background(), gophkeeper.Blob{}, "")
		assert.NotNil(t, storeBlobError)

		_, restorePieceError := identity.RestorePiece(context.Background(), 0, "")
		assert.NotNil(t, restorePieceError)

		_, restoreBlobError := identity.RestoreBlob(context.Background(), 0, "")
		assert.NotNil(t, restoreBlobError)
	})
}
