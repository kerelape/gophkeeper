package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"

	"github.com/kerelape/gophkeeper/internal/cli"
)

func main() {
	var server = flag.String("s", "", "Gophkeeper address")
	flag.Parse()
	if *server == "" {
		log.Fatal("missing -s flag")
	}

	var application = cli.CLI{
		Gophkeeper: &gophkeeper.RestGophkeeper{
			Server: *server,
			Client: *http.DefaultClient,
		},
	}
	if err := application.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
