package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/kerelape/gophkeeper/internal/cli"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/rest"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("")
	server := flag.String("s", "", "Gophkeeper address")
	flag.Parse()
	if *server == "" {
		log.Fatal("missing -s flag")
	}

	application := cli.CLI{
		Gophkeeper: &rest.Gophkeeper{
			Server: *server,
			Client: http.Client{},
		},
		CommandLine: flag.Args(),
	}
	if err := application.Run(context.Background()); err != nil {
		log.Println()
		log.Fatal(err)
	}
}
