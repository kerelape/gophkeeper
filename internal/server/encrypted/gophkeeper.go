package encrypted

import (
	"context"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// Gophkeeper is an ecrypted Gophkeeper.
type Gophkeeper struct {
	Origin gophkeeper.Gophkeeper
	Cipher Cipher
}

var _ gophkeeper.Gophkeeper = (*Gophkeeper)(nil)

// Authenticate implements gophkeeper.Gophkeeper.
func (g Gophkeeper) Authenticate(ctx context.Context, credential gophkeeper.Credential) (gophkeeper.Token, error) {
	return g.Origin.Authenticate(ctx, credential)
}

// Identity implements gophkeeper.Gophkeeper.
func (g Gophkeeper) Identity(ctx context.Context, token gophkeeper.Token) (gophkeeper.Identity, error) {
	origin, originError := g.Origin.Identity(ctx, token)
	identity := Identity{
		Origin: origin,
		Cipher: g.Cipher,
	}
	return identity, originError
}

// Register implements gophkeeper.Gophkeeper.
func (g Gophkeeper) Register(ctx context.Context, credential gophkeeper.Credential) error {
	return g.Origin.Register(ctx, credential)
}
