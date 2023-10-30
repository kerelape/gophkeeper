package server

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// UsernameBasedTokenSource is a token provider based on username.
type UsernameBasedTokenSource = TokenSource[string]

type (
	jwtSource struct {
		secret   []byte
		lifespan time.Duration
	}
)

var _ UsernameBasedTokenSource = (*jwtSource)(nil)

// NewJWTSource creates a new JWT provider.
func NewJWTSource(secret []byte, lifespan time.Duration) UsernameBasedTokenSource {
	return &jwtSource{
		secret:   secret,
		lifespan: lifespan,
	}
}

// Create implements Provider.
func (jp *jwtSource) Create(_ context.Context, username string) (gophkeeper.Token, error) {
	rawToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"exp": time.Now().Add(jp.lifespan).Unix(),
			"sub": username,
		},
	)
	token, signTokenError := rawToken.SignedString(jp.secret)
	return (gophkeeper.Token)(token), signTokenError
}

// Unwrap implements Provider.
func (jp *jwtSource) Unwrap(_ context.Context, token gophkeeper.Token) (string, error) {
	parsedToken, parseTokenError := jwt.Parse(
		(string)(token),
		func(t *jwt.Token) (interface{}, error) {
			return jp.secret, nil
		},
	)
	if parseTokenError != nil {
		return "", errors.Join(parseTokenError, gophkeeper.ErrBadCredential)
	}

	return parsedToken.Claims.GetSubject()
}
