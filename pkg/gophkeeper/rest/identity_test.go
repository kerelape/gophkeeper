package rest_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
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

	registerError := g.Register(context.Background(), credential)
	assert.Nil(t, registerError, "did not expect an error")

	token, authenticateError := g.Authenticate(context.Background(), credential)
	assert.Nil(t, authenticateError, "did not expect an error")

	identity, identityError := g.Identity(context.Background(), token)
	assert.Nil(t, identityError, "did not expect an error")

	var rescount int

	t.Run("Piece", func(t *testing.T) {
		var rid gophkeeper.ResourceID
		t.Run("Store", func(t *testing.T) {
			r, err := identity.StorePiece(
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
				_, err := identity.StorePiece(
					context.Background(),
					gophkeeper.Piece{},
					"_",
				)
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrBadCredential, "unexpected error")
			})
		})
		t.Run("Restore", func(t *testing.T) {
			piece, err := identity.RestorePiece(
				context.Background(),
				rid,
				credential.Password,
			)
			assert.Nil(t, err, "did not expect an error")
			assert.Equal(t, "testmeta", piece.Meta, "restored meta incorrectly")
			assert.Equal(t, "testcontent", (string)(piece.Content), "restored content incorrectly")
			t.Run("Invalid password", func(t *testing.T) {
				_, err := identity.RestorePiece(context.Background(), rid, "_")
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrBadCredential, "unexpected error")
			})
			t.Run("Invalid RID", func(t *testing.T) {
				_, err := identity.RestorePiece(context.Background(), -1, credential.Password)
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
			})
		})
	})
	t.Run("Blob", func(t *testing.T) {
		var rid gophkeeper.ResourceID
		t.Run("Store", func(t *testing.T) {
			r, err := identity.StoreBlob(
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
				_, err := identity.StoreBlob(
					context.Background(),
					gophkeeper.Blob{},
					"_",
				)
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrBadCredential, "unexpected error")
			})
		})
		t.Run("Restore", func(t *testing.T) {
			piece, err := identity.RestoreBlob(
				context.Background(),
				rid,
				credential.Password,
			)
			assert.Nil(t, err, "did not expect an error")
			assert.Equal(t, "testmeta", piece.Meta, "restored meta incorrectly")
			content, contentError := io.ReadAll(piece.Content)
			assert.Nil(t, contentError, "did not expect an error")
			assert.Equal(t, "testcontent", (string)(content), "restored content incorrectly")
			t.Run("Invalid password", func(t *testing.T) {
				_, err := identity.RestoreBlob(context.Background(), rid, "_")
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrBadCredential, "unexpected error")
			})
			t.Run("Invalid RID", func(t *testing.T) {
				_, err := identity.RestoreBlob(context.Background(), -1, credential.Password)
				assert.NotNil(t, err, "expected an error")
				assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
			})
		})
	})
	t.Run("Delete", func(t *testing.T) {
		assert.Positive(t, rescount, "expected rescount to grow by this moment")
		err := identity.Delete(context.Background(), (gophkeeper.ResourceID)(rescount-1))
		assert.Nil(t, err, "did not expect an error")
		rescount--
		t.Run("Invalid RID", func(t *testing.T) {
			err := identity.Delete(context.Background(), -1)
			assert.NotNil(t, err, "expected an error")
			assert.ErrorIs(t, err, gophkeeper.ErrResourceNotFound, "unexpected error")
		})
	})
	t.Run("List", func(t *testing.T) {
		resources, err := identity.List(context.Background())
		assert.Nil(t, err, "did not expect an error")
		assert.Equal(t, rescount, len(resources), "incorrect resources count")
	})

	t.Run("Invalid configuration", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		identity := rest.Identity{
			Client: *server.Client(),
			Server: server.URL,
			Token:  "",
		}
		_, storePieceError := identity.StorePiece(context.Background(), gophkeeper.Piece{}, "")
		assert.NotNil(t, storePieceError)
		_, restorePieceError := identity.RestorePiece(context.Background(), 0, "")
		assert.NotNil(t, restorePieceError)
		_, storeBlobError := identity.StoreBlob(
			context.Background(),
			gophkeeper.Blob{Content: io.NopCloser(bytes.NewReader([]byte{}))},
			"",
		)
		assert.NotNil(t, storeBlobError)
		_, restoreBlobError := identity.RestoreBlob(context.Background(), 0, "")
		assert.NotNil(t, restoreBlobError)
		_, listError := identity.List(context.Background())
		assert.NotNil(t, listError)
		deleteError := identity.Delete(context.Background(), 0)
		assert.NotNil(t, deleteError)
	})
	t.Run("nil context", func(t *testing.T) {
		_, storePieceError := identity.StorePiece(nilContext, gophkeeper.Piece{}, credential.Password)
		assert.NotNil(t, storePieceError)
		_, storeBlobError := identity.StoreBlob(
			nilContext,
			gophkeeper.Blob{Content: io.NopCloser(bytes.NewReader([]byte{}))},
			credential.Password,
		)
		assert.NotNil(t, storeBlobError)
		_, restorePieceError := identity.RestorePiece(nilContext, 0, credential.Password)
		assert.NotNil(t, restorePieceError)
		_, restoreBlobError := identity.RestoreBlob(nilContext, 0, credential.Password)
		assert.NotNil(t, restoreBlobError)
		_, listError := identity.List(nilContext)
		assert.NotNil(t, listError)
		deleteError := identity.Delete(nilContext, 0)
		assert.NotNil(t, deleteError)
	})
}
