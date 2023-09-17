package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	serverrest "github.com/kerelape/gophkeeper/internal/server/rest"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/rest"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/virtual"
	"github.com/stretchr/testify/assert"
)

func GophkeeperExample() {
	var g = rest.Gophkeeper{
		Client: *http.DefaultClient,       // HTTP client to be used by the object.
		Server: "https://localhost:16355", // Address of the REST server.
	}

	// Credentials to register and then authenticate the new user.
	var credential = gophkeeper.Credential{
		Username: "gophuser",
		Password: "querty",
	}

	// Register the new user.
	g.Register(context.Background(), credential)

	// Authenticate the user with the same credentials.
	var token, _ = g.Authenticate(context.Background(), credential)

	// User authentication token to get a REST identity.
	var identity, _ = g.Identity(context.Background(), token)

	// Piece to be stored by Gophkeeper.
	var piece = gophkeeper.Piece{
		Meta:    "This meta information won't get encypted by Gophkeeper",
		Content: ([]byte)("This WILL get encrypted and securely stored by Gophkeeper."),
	}

	// Store the piece and get its RID back.
	var rid, _ = identity.StorePiece(context.Background(), piece, credential.Password)

	// Restore the piece back using its RID.
	piece, _ = identity.RestorePiece(context.Background(), rid, credential.Password)
}

func TestGophkeeper(t *testing.T) {
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
	t.Run("Register", func(t *testing.T) {
		var err = g.Register(context.Background(), credential)
		assert.Nil(t, err, "did not expect an error")
		t.Run("Subsequent", func(t *testing.T) {
			var err = g.Register(context.Background(), credential)
			assert.NotNil(t, err, "expected an error")
			assert.ErrorIs(t, err, gophkeeper.ErrIdentityDuplicate, "unexpected error")
		})
	})
	t.Run("Authenticate", func(t *testing.T) {
		var _, err = g.Authenticate(context.Background(), credential)
		assert.Nil(t, err, "did not expect an error")
		t.Run("Invalid credential", func(t *testing.T) {
			var _, err = g.Authenticate(context.Background(), gophkeeper.Credential{})
			assert.NotNil(t, err, "expected an error")
			assert.ErrorIs(t, err, gophkeeper.ErrBadCredential)
		})
	})
}
