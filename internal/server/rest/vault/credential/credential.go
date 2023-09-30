// Package credential provides authorization related
// middleware and function.
package credential

import (
	"context"
	"net/http"
)

type contextKey string

const contextKeyPassword = contextKey("password")

// Middleware is vault credential middleware.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(out http.ResponseWriter, in *http.Request) {
		password := in.Header.Get("X-Password")
		if password == "" {
			status := http.StatusBadRequest
			http.Error(out, http.StatusText(status), status)
			return
		}
		next.ServeHTTP(
			out,
			in.WithContext(
				context.WithValue(
					in.Context(),
					contextKeyPassword,
					password,
				),
			),
		)
	})
}

// Password returns vault password.
func Password(in *http.Request) string {
	password, ok := in.Context().Value(contextKeyPassword).(string)
	if !ok {
		panic("unexpected value type (use middlware)")
	}
	return password
}
