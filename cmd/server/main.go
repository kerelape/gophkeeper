package main

import (
	"encoding/base64"
	"log"
	"os"
	"path"

	"github.com/kerelape/gophkeeper/cmd/server/config"
	"github.com/kerelape/gophkeeper/internal/server"
	"github.com/pior/runnable"
)

func main() {
	log.SetPrefix("[GOPHKEEPER] ")
	var configuration config.Config
	if err := config.Read(&configuration); err != nil {
		log.Fatal(configuration.Description())
	}
	secret, decodeSecretError := base64.RawStdEncoding.DecodeString(configuration.Token.Secret)
	if decodeSecretError != nil {
		log.Fatalf("failed to parse token secret: %s", decodeSecretError.Error())
	}
	wd, wdError := os.Getwd()
	if wdError != nil {
		log.Fatalf(wdError.Error())
	}
	gophkeeper := server.Server{
		RestAddress:       configuration.Rest.Address,
		RestUseTLS:        configuration.Rest.UseTLS,
		RestHostWhilelist: configuration.Rest.HostWhilelist,

		DatabaseDSN: configuration.DatabaseDSN,
		BlobsDir:    path.Join(wd, "blobs"),

		TokenSecret:   secret,
		TokenLifespan: configuration.Token.Lifespan,

		UsernameMinLength: configuration.UsernameMinLength,
		PasswordMinLength: configuration.PasswordMinLength,
	}
	runnable.Run(&gophkeeper)
}
