package rest_test

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	serverrest "github.com/kerelape/gophkeeper/internal/server/rest"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/rest"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/virtual"
	"github.com/stretchr/testify/assert"
)

func TestIdentity(t *testing.T) {
	var (
		entry = serverrest.Entry{
			Gophkeeper: virtual.New(
				time.Hour,
				t.TempDir(),
			),
		}
		server = httptest.NewServer(entry.Route())
	)
	var (
		client = server.Client()
		g      = rest.Gophkeeper{
			Client: *client,
			Server: server.URL,
		}
		credential = gophkeeper.Credential{
			Username: "test",
			Password: "qwerty",
		}
	)
	defer server.Close()

	var registerError = g.Register(context.Background(), credential)
	assert.Nil(t, registerError, "did not expect and error")

	var token, authenticateError = g.Authenticate(context.Background(), credential)
	assert.Nil(t, authenticateError, "did not expect and error")

	var identity, identityError = g.Identity(context.Background(), token)
	assert.Nil(t, identityError, "did not expect and error")

	var rescount int

	t.Run("Piece", func(t *testing.T) {
		var rid gophkeeper.ResourceID
		t.Run("Store", func(t *testing.T) {
			var r, err = identity.StorePiece(
				context.Background(),
				gophkeeper.Piece{
					Meta:    "testmeta",
					Content: ([]byte)("testcontent"),
				},
				credential.Password,
			)
			assert.Nil(t, err, "did not expect an error")
			rid = r
			rescount++
			t.Run("Invalid password", func(t *testing.T) {
				var _, err = identity.StorePiece(
					context.Background(),
					gophkeeper.Piece{},
					"_",
				)
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrBadCredential, "unexpected error")
			})
		})
		t.Run("Restore", func(t *testing.T) {
			var piece, err = identity.RestorePiece(
				context.Background(),
				rid,
				credential.Password,
			)
			assert.Nil(t, err, "did not expect an error")
			assert.Equal(t, "testmeta", piece.Meta, "restored meta incorrectly")
			assert.Equal(t, "testcontent", (string)(piece.Content), "restored content incorrectly")
			t.Run("Invalid password", func(t *testing.T) {
				var _, err = identity.RestorePiece(context.Background(), rid, "_")
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrBadCredential, "unexpected error")
			})
			t.Run("Invalid RID", func(t *testing.T) {
				var _, err = identity.RestorePiece(context.Background(), -1, credential.Password)
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
			})
		})
	})
	t.Run("Blob", func(t *testing.T) {
		var rid gophkeeper.ResourceID
		t.Run("Store", func(t *testing.T) {
			var r, err = identity.StoreBlob(
				context.Background(),
				gophkeeper.Blob{
					Meta:    "testmeta",
					Content: io.NopCloser(strings.NewReader("testcontent")),
				},
				credential.Password,
			)
			assert.Nil(t, err, "did not expect an error")
			rid = r
			rescount++
			t.Run("Invalid password", func(t *testing.T) {
				var _, err = identity.StoreBlob(
					context.Background(),
					gophkeeper.Blob{},
					"_",
				)
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrBadCredential, "unexpected error")
			})
		})
		t.Run("Restore", func(t *testing.T) {
			var piece, err = identity.RestoreBlob(
				context.Background(),
				rid,
				credential.Password,
			)
			assert.Nil(t, err, "did not expect an error")
			assert.Equal(t, "testmeta", piece.Meta, "restored meta incorrectly")
			var content, contentError = io.ReadAll(piece.Content)
			assert.Nil(t, contentError, "did not expect an error")
			assert.Equal(t, "testcontent", (string)(content), "restored content incorrectly")
			t.Run("Invalid password", func(t *testing.T) {
				var _, err = identity.RestoreBlob(context.Background(), rid, "_")
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrBadCredential, "unexpected error")
			})
			t.Run("Invalid RID", func(t *testing.T) {
				var _, err = identity.RestoreBlob(context.Background(), -1, credential.Password)
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		assert.Positive(t, rescount, "expected rescount to grow by this moment")
		var err = identity.Delete(context.Background(), (gophkeeper.ResourceID)(rescount-1))
		assert.Nil(t, err, "did not expect and error")
		rescount--
		t.Run("Invalid RID", func(t *testing.T) {
			var err = identity.Delete(context.Background(), -1)
			assert.NotNil(t, err, "expected an error")
			assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
		})
	})
	t.Run("List", func(t *testing.T) {
		var resources, err = identity.List(context.Background())
		assert.Nil(t, err, "did not expect an error")
		assert.Equal(t, rescount, len(resources), "incorrect resources count")
	})
}
