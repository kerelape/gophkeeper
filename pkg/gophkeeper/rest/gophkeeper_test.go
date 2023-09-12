package rest

import (
	"context"
	"net/http"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

func GophkeeperExample() {
	var g = Gophkeeper{
		Client: *http.DefaultClient,       // HTTP client to be used by the object.
		Server: "https://localhost:16355", // Address of the REST server.
	}

	// Credentials to register and then authenticate the new user.
	var credential = gophkeeper.Credential{
		Username: "gophuser",
		Password: "querty",
	}

	// Register the new user.
	g.Register(context.Background(), credential)

	// Authenticate the user with the same credentials.
	var token, _ = g.Authenticate(context.Background(), credential)

	// User authentication token to get a REST identity.
	var identity, _ = g.Identity(context.Background(), token)

	// Piece to be stored by Gophkeeper.
	var piece = gophkeeper.Piece{
		Meta:    "This meta information won't get encypted by Gophkeeper",
		Content: ([]byte)("This WILL get encrypted and securely stored by Gophkeeper."),
	}

	// Store the piece and get its RID back.
	var rid, _ = identity.StorePiece(context.Background(), piece, credential.Password)

	// Restore the piece back using its RID.
	piece, _ = identity.RestorePiece(context.Background(), rid, credential.Password)
}
