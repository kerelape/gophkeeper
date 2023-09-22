package authentication

import (
	"context"
	"errors"
	"net/http"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

type contextKey string

const contextKeyIdentity = (contextKey)("identity")

// Middleware is authentication middleware.
func Middleware(g gophkeeper.Gophkeeper) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(out http.ResponseWriter, in *http.Request) {
			token := in.Header.Get("Authorization")
			if token == "" {
				status := http.StatusUnauthorized
				http.Error(out, http.StatusText(status), status)
				return
			}
			identity, identityError := g.Identity(in.Context(), (gophkeeper.Token)(token))
			if identityError != nil {
				status := http.StatusInternalServerError
				if errors.Is(identityError, gophkeeper.ErrBadCredential) {
					status = http.StatusUnauthorized
				}
				http.Error(out, http.StatusText(status), status)
				return
			}
			next.ServeHTTP(out, in.WithContext(context.WithValue(in.Context(), contextKeyIdentity, identity)))
		})
	}
}

// Identity returns identity assigned to the request.
func Identity(in *http.Request) gophkeeper.Identity {
	identity := in.Context().Value(contextKeyIdentity)
	if identity == nil {
		panic("identity is not set (use middlware)")
	}
	if identity, ok := identity.(gophkeeper.Identity); ok {
		return identity
	}
	panic("identity is set incorrectly")
}
