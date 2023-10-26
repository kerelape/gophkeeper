package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTSource(t *testing.T) {
	jp := NewJWTSource(([]byte)("secret"), time.Hour)

	token, tokenError := jp.Create(context.Background(), "test")
	assert.Nil(t, tokenError, "did not expect an error")

	username, usernameError := jp.Unwrap(context.Background(), token)
	assert.Nil(t, usernameError, "did not expect an error")
	assert.Equal(t, username, "test", "usernames do not match")

	t.Run("Expiration", func(t *testing.T) {
		jp := NewJWTSource(([]byte)("secret"), time.Second)

		token, tokenError := jp.Create(context.Background(), "test")
		assert.Nil(t, tokenError, "did not expect an error")

		time.Sleep(time.Second * 2)

		_, err := jp.Unwrap(context.Background(), token)
		assert.NotNil(t, err, "expected to get an error")
	})
}
