package main

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/kerelape/gophkeeper/cmd/server/config"
	"github.com/kerelape/gophkeeper/internal/server"
	"github.com/kerelape/gophkeeper/internal/server/postgres"
	"github.com/kerelape/gophkeeper/internal/server/rest"
	"github.com/pior/runnable"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	var configuration config.Config
	if err := config.Read(&configuration); err != nil {
		log.Fatal(configuration.Description())
	}

	log.SetPrefix("[GOPHKEEPER] ")
	secret, decodeSecretError := base64.RawStdEncoding.DecodeString(configuration.Token.Secret)
	if decodeSecretError != nil {
		log.Fatalf("failed to parse token secret: %s", decodeSecretError.Error())
	}

	wd, wdError := os.Getwd()
	if wdError != nil {
		log.Fatalf(wdError.Error())
	}

	var (
		database = postgres.New(
			postgres.DSNSource(configuration.DatabaseDSN),
			server.NewJWTSource(
				secret,
				configuration.Token.Lifespan,
			),
			postgres.WithBlobsDir(path.Join(wd, "blobs")),
			postgres.WithPasswordEncoding(base64.RawStdEncoding),
		)
		rst = rest.Entry{
			Gophkeeper: database,
		}
		srv = http.Server{
			Addr:    configuration.Rest.Address,
			Handler: rst.Route(),
		}
	)

	manager := runnable.NewManager()
	manager.Add(database)

	if configuration.Rest.UseTLS {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(configuration.Rest.HostWhitelist...),
		}
		srv.TLSConfig = m.TLSConfig()
		manager.Add(runnable.Func(func(ctx context.Context) error {
			ch := make(chan error)
			go func() {
				ch <- srv.ListenAndServeTLS("", "")
			}()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err := <-ch:
				return err
			}
		}))
	} else {
		manager.Add(runnable.HTTPServer(&srv))
	}

	runnable.Run(manager.Build())
}
