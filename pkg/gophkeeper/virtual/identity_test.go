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
	var registerError = g.Register(context.Background(), credential)
	assert.Nil(t, registerError, "expected to successfully register")

	var registerAlianError = g.Register(context.Background(), alianCredential)
	assert.Nil(t, registerAlianError, "expected to successfully register")

	var token, authenticateError = g.Authenticate(context.Background(), credential)
	assert.Nil(t, authenticateError, "expected to successfully authenticate")

	var alianToken, authenticateAlianError = g.Authenticate(context.Background(), alianCredential)
	assert.Nil(t, authenticateAlianError, "expected to successfully authenticate")

	var identity, identityError = g.Identity(context.Background(), token)
	assert.Nil(t, identityError, "expected to successfully get the identity")

	var alian, alianError = g.Identity(context.Background(), alianToken)
	assert.Nil(t, alianError, "expected to successfully get the identity")

	var rid, storeError = identity.StorePiece(
		context.Background(),
		gophkeeper.Piece{
			Meta:    "testmeta",
			Content: ([]byte)("testcontent"),
		},
		credential.Password,
	)
	assert.Nil(t, storeError, "expected to successfully store a piece")
	var piece, restoreError = identity.RestorePiece(
		context.Background(),
		rid, credential.Password,
	)
	assert.Nil(t, restoreError, "expected to successfully restore the piece back")
	assert.Equal(t, "testmeta", piece.Meta, "meta is not restored correctly")
	assert.Equal(t, "testcontent", (string)(piece.Content), "content is not restored correctly")

	var _, alianRestoreError = alian.RestorePiece(
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

	var blobRID, storeBlobError = identity.StoreBlob(
		context.Background(),
		gophkeeper.Blob{
			Meta:    "testmeta",
			Content: io.NopCloser(bytes.NewReader(([]byte)("testcontent"))),
		},
		credential.Password,
	)
	assert.Nil(t, storeBlobError, "expected to successfully store a blob")

	var blob, restoreBlobError = identity.RestoreBlob(
		context.Background(),
		blobRID, credential.Password,
	)
	assert.Nil(t, restoreBlobError, "expected to successfully restore a blob")
	assert.Equal(t, "testmeta", blob.Meta, "meta is not restored correctly")
	assert.Equal(
		t,
		"testcontent",
		func() string {
			var content, err = io.ReadAll(blob.Content)
			assert.Nil(t, err, "expected to successfully read content")
			return (string)(content)
		}(),
		"content is not restored correctly",
	)
	assert.Nil(t, blob.Content.Close(), "failed to close blob content")

	var _, restoreAlianBlob = alian.RestoreBlob(
		context.Background(),
		blobRID, alianCredential.Password,
	)
	assert.NotNil(t, restoreAlianBlob, "expected to get an error on restoring alian blob")
	assert.ErrorIs(t, restoreAlianBlob, gophkeeper.ErrResourceNotFound, "unexpected error")

	var deleteError = identity.Delete(context.Background(), rid)
	assert.Nil(t, deleteError, "did not expect an error")

	var alianDeleteError = alian.Delete(context.Background(), blobRID)
	assert.NotNil(t, alianDeleteError, "expected an error")
	assert.ErrorIs(t, alianDeleteError, gophkeeper.ErrResourceNotFound, "unexpected error")

	var resources, listError = identity.List(context.Background())
	assert.Nil(t, listError, "did not expect an error")
	assert.Equal(t, 1, len(resources), "expected resources list to be equal to 2")
}
