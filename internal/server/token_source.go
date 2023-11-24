// Package token provides everything needed for token creation.
package server

import (
	"context"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// TokenSource is a token creation unit.
type TokenSource[Credentials any] interface {
	// Create creates a token for the given credentials.
	Create(context.Context, Credentials) (gophkeeper.Token, error)

	// Unwrap returns the original credentials stored by the token.
	Unwrap(context.Context, gophkeeper.Token) (Credentials, error)
}
