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
	if signTokenError != nil {
		return gophkeeper.InvalidToken, signTokenError
	}

	return (gophkeeper.Token)(token), nil
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
		return "", parseTokenError
	}

	claims := parsedToken.Claims

	exp, expError := claims.GetExpirationTime()
	if expError != nil {
		return "", expError
	}
	if exp.Before(time.Now()) {
		return "", errors.New("token expired")
	}

	sub, subError := claims.GetSubject()
	if subError != nil {
		return "", subError
	}

	return sub, nil
}
