package rest_test

import (
	"context"
	"fmt"
	"log"
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

var nilContext context.Context = nil

func GophkeeperExample() {
	g := rest.Gophkeeper{
		Client: *http.DefaultClient,       // HTTP client to be used by the object.
		Server: "https://localhost:16355", // Address of the REST server.
	}

	// Credentials to register and then authenticate the new user.
	credential := gophkeeper.Credential{
		Username: "gophuser",
		Password: "querty",
	}

	// Register the new user.
	err := g.Register(context.Background(), credential)
	if err != nil {
		log.Fatal("We failed to register!")
	}

	// Authenticate the user with the same credentials.
	token, _ := g.Authenticate(context.Background(), credential)

	// User authentication token to get a REST identity.
	identity, _ := g.Identity(context.Background(), token)

	// Piece to be stored by Gophkeeper.
	piece := gophkeeper.Piece{
		Meta:    "This meta information won't get encypted by Gophkeeper",
		Content: ([]byte)("This WILL get encrypted and securely stored by Gophkeeper."),
	}

	// Store the piece and get its RID back.
	rid, _ := identity.StorePiece(context.Background(), piece, credential.Password)

	// Restore the piece back using its RID.
	piece, _ = identity.RestorePiece(context.Background(), rid, credential.Password)
	fmt.Println(piece.Meta) // -> This meta information won't get encypted by Gophkeeper
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
		err := g.Register(context.Background(), credential)
		assert.Nil(t, err, "did not expect an error")
		t.Run("Subsequent", func(t *testing.T) {
			err := g.Register(context.Background(), credential)
			assert.NotNil(t, err, "expected an error")
			assert.ErrorIs(t, err, gophkeeper.ErrIdentityDuplicate, "unexpected error")
		})

		t.Run("nil context", func(t *testing.T) {
			err := g.Register(nilContext, credential)
			assert.NotNil(t, err)
		})

		t.Run("invalid server url", func(t *testing.T) {
			g := g
			g.Server = ""
			err := g.Register(context.Background(), credential)
			assert.NotNil(t, err)
		})
	})
	t.Run("Authenticate", func(t *testing.T) {
		_, err := g.Authenticate(context.Background(), credential)
		assert.Nil(t, err, "did not expect an error")
		t.Run("Invalid credential", func(t *testing.T) {
			_, err := g.Authenticate(context.Background(), gophkeeper.Credential{})
			assert.NotNil(t, err, "expected an error")
			assert.ErrorIs(t, err, gophkeeper.ErrBadCredential)
		})

		t.Run("nil context", func(t *testing.T) {
			_, err := g.Authenticate(nilContext, credential)
			assert.NotNil(t, err)
		})

		t.Run("invalid server url", func(t *testing.T) {
			g := g
			g.Server = ""
			_, err := g.Authenticate(context.Background(), credential)
			assert.NotNil(t, err)
		})
	})

	t.Run("Unexpected status code", func(t *testing.T) {
		var (
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
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

		t.Run("Register", func(t *testing.T) {
			err := g.Register(context.Background(), credential)
			assert.NotNil(t, err)
		})
		t.Run("Authenticate", func(t *testing.T) {
			_, err := g.Authenticate(context.Background(), credential)
			assert.NotNil(t, err)
		})
	})
}
