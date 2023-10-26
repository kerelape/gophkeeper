// Package server provides Server object.
package server

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/kerelape/gophkeeper/internal/server/encrypted"
	"github.com/kerelape/gophkeeper/internal/server/postgres"
	"github.com/kerelape/gophkeeper/internal/server/rest"
	"github.com/pior/runnable"
)

// Server is gophkeeper server.
type Server struct {
	RestAddress       string // the address that REST api serves at.
	RestUseTLS        bool
	RestHostWhitelist []string

	DatabaseDSN   string
	BlobsDir      string
	TokenSecret   []byte
	TokenLifespan time.Duration

	UsernameMinLength uint
	PasswordMinLength uint
}

var _ runnable.Runnable = (*Server)(nil)

// Run runs Server.
func (s *Server) Run(ctx context.Context) error {
	var (
		database = postgres.Gophkeeper{
			PasswordEncoding: base64.RawStdEncoding,

			Source:   (postgres.DSNSource)(s.DatabaseDSN),
			BlobsDir: s.BlobsDir,

			TokenSecret:   s.TokenSecret,
			TokenLifespan: s.TokenLifespan,

			UsernameMinLength: s.UsernameMinLength,
			PasswordMinLength: s.PasswordMinLength,
		}
		restDaemon = rest.Rest{
			Address: s.RestAddress,
			Gophkeeper: encrypted.Gophkeeper{
				Origin: &database,
				Cipher: encrypted.CFBCipher{},
			},
			UseTLS:        s.RestUseTLS,
			HostWhilelist: s.RestHostWhilelist,
		}
	)

	manager := runnable.NewManager()
	manager.Add(&database)
	manager.Add(&restDaemon)
	return manager.Build().Run(ctx)
}
