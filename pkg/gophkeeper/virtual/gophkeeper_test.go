package virtual_test

import (
	"context"
	"testing"
	"time"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/virtual"
	"github.com/stretchr/testify/assert"
)

func TestGophkeeper(t *testing.T) {
	t.Run("Registration", func(t *testing.T) {
		var (
			g          = virtual.New(time.Hour, t.TempDir())
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)
		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")
	})
	t.Run("Subsequent registration with same credential", func(t *testing.T) {
		var (
			g          = virtual.New(time.Hour, t.TempDir())
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)
		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		reregisterError := g.Register(context.Background(), credential)
		assert.NotNil(
			t,
			reregisterError,
			"expected to get an error on reregister attempt",
		)
		assert.ErrorIs(
			t,
			reregisterError,
			gophkeeper.ErrIdentityDuplicate,
			"unexpected error on reregister",
		)
	})
	t.Run("Authentication", func(t *testing.T) {
		var (
			g          = virtual.New(time.Hour, t.TempDir())
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)
		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		token, authenticateError := g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authentiate")

		_, identityError := g.Identity(context.Background(), token)
		assert.Nil(t, identityError, "expected to successfully get identity")
	})
	t.Run("Authentication with invalid credential", func(t *testing.T) {
		var (
			g          = virtual.New(time.Hour, t.TempDir())
			credential = gophkeeper.Credential{}
		)
		_, err := g.Authenticate(context.Background(), credential)
		assert.NotNil(
			t,
			err,
			"expected to get an error on authetntication with invalid credential",
		)
		assert.ErrorIs(
			t,
			err,
			gophkeeper.ErrBadCredential,
			"unexpected error on authentication with invalid credential",
		)
	})
	t.Run("Token expiration", func(t *testing.T) {
		const timeout = time.Second
		var (
			g          = virtual.New(timeout, t.TempDir())
			credential = gophkeeper.Credential{
				Username: "test",
				Password: "qwerty",
			}
		)
		registerError := g.Register(context.Background(), credential)
		assert.Nil(t, registerError, "expected to successfully register")

		token, authenticateError := g.Authenticate(context.Background(), credential)
		assert.Nil(t, authenticateError, "expected to successfully authentiate")

		time.Sleep(timeout + time.Second)

		_, identityError := g.Identity(context.Background(), token)
		assert.NotNil(t, identityError, "expected to get an error with an outdated token")
		assert.ErrorIs(
			t,
			identityError,
			gophkeeper.ErrBadCredential,
			"unexpected error on identity retrivation with an outdated token",
		)
	})
}
