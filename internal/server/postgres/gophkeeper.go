package postgres

import (
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"io/fs"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kerelape/gophkeeper/internal/deferred"
	"github.com/kerelape/gophkeeper/internal/server"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
	"github.com/pior/runnable"
	"golang.org/x/crypto/bcrypt"
)

//go:embed init.sql
var initQuery string

type (
	// Gophkeeper is a postgresql identity repository.
	Gophkeeper struct {
		passwordEncoding *base64.Encoding
		source           DatabaseSource
		blobsDir         string
		tokenSource      server.UsernameBasedTokenSource

		connection deferred.Deferred[*pgx.Conn]
	}
	option func(g *Gophkeeper)
)

// New craetes a new postgres Gophkeeper and returns it.
func New(source DatabaseSource, tokenSource server.UsernameBasedTokenSource, options ...option) *Gophkeeper {
	g := &Gophkeeper{
		passwordEncoding: base64.RawStdEncoding,
		source:           source,
		tokenSource:      tokenSource,
		blobsDir:         "./blobs",
	}
	for _, o := range options {
		o(g)
	}
	return g
}

var (
	_ gophkeeper.Gophkeeper = (*Gophkeeper)(nil)
	_ runnable.Runnable     = (*Gophkeeper)(nil)
)

// Register implements Repository.
func (r *Gophkeeper) Register(ctx context.Context, credential gophkeeper.Credential) error {
	connection, connectionError := r.connection.Get(ctx)
	if connectionError != nil {
		return connectionError
	}

	password, passwordError := bcrypt.GenerateFromPassword(
		([]byte)(credential.Password),
		bcrypt.DefaultCost,
	)
	if passwordError != nil {
		return passwordError
	}

	_, insertError := connection.Exec(
		ctx,
		`INSERT INTO identities(username, password) VALUES($1, $2)`,
		credential.Username,
		r.passwordEncoding.EncodeToString(password),
	)
	if insertError != nil {
		if err := new(pgconn.PgError); errors.As(insertError, &err) && err.Code == "23505" {
			return gophkeeper.ErrIdentityDuplicate
		}
		return insertError
	}

	return nil
}

// Authenticate implements Repository.
func (r *Gophkeeper) Authenticate(ctx context.Context, credential gophkeeper.Credential) (gophkeeper.Token, error) {
	connection, connectionError := r.connection.Get(ctx)
	if connectionError != nil {
		return gophkeeper.InvalidToken, connectionError
	}

	identity := Identity{
		Connection:       connection,
		PasswordEncoding: r.passwordEncoding,
		Username:         credential.Username,
	}
	if err := identity.comparePassword(ctx, credential.Password); err != nil {
		return gophkeeper.InvalidToken, err
	}

	t, tError := r.tokenSource.Create(ctx, credential.Username)
	if tError != nil {
		return gophkeeper.InvalidToken, tError
	}

	return (gophkeeper.Token)(t), nil
}

// Identity implements Repository.
func (r *Gophkeeper) Identity(ctx context.Context, token gophkeeper.Token) (gophkeeper.Identity, error) {
	connection, connectionError := r.connection.Get(ctx)
	if connectionError != nil {
		return nil, connectionError
	}

	username, usernameError := r.tokenSource.Unwrap(ctx, token)
	if usernameError != nil {
		return nil, errors.Join(usernameError, gophkeeper.ErrBadCredential)
	}

	identity := &Identity{
		Connection:       connection,
		PasswordEncoding: r.passwordEncoding,
		Username:         username,
		BlobsDir:         r.blobsDir,
	}
	return identity, nil
}

// Run implements Runnable.
func (r *Gophkeeper) Run(ctx context.Context) error {
	mkdirError := os.MkdirAll(r.blobsDir, fs.ModePerm)
	if mkdirError != nil {
		return mkdirError
	}

	connection, connectError := r.source.Connect(ctx)
	if connectError != nil {
		return connectError
	}
	defer connection.Close(context.Background())

	_, initializeError := connection.Exec(ctx, initQuery)
	if initializeError != nil {
		return initializeError
	}

	r.connection.Set(connection)

	<-ctx.Done()
	return ctx.Err()
}

// WithBlobsDir sets blobs dir to the gophkeeper.
func WithBlobsDir(dir string) option {
	return func(g *Gophkeeper) {
		g.blobsDir = dir
	}
}

// WithPasswordEnoding sets password encoding to the gophkeeper.
func WithPasswordEncoding(encoding *base64.Encoding) option {
	if encoding == nil {
		panic("encoding must be not nil")
	}
	return func(g *Gophkeeper) {
		g.passwordEncoding = encoding
	}
}
