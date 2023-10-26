package virtual

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kerelape/gophkeeper/internal/server"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

const invalidIdentityID = -1

type identity struct {
	username string
	password string
}

// Gophkeeper is a virtual Gophkeeper.
type Gophkeeper struct {
	identities []identity

	ts       server.UsernameBasedTokenSource
	blobsDir string

	storage *storage

	mutex *sync.Mutex
}

// New returns a new virtual Gophkeeper that store all its
// data in RAM.
func New(sessionLifespan time.Duration, blobsDir string) *Gophkeeper {
	return &Gophkeeper{
		identities: make([]identity, 0),
		ts:         server.NewJWTSource(([]byte)("none"), sessionLifespan),
		blobsDir:   blobsDir,
		storage: &storage{
			mutex:     &sync.Mutex{},
			resources: make([]resource, 0),
			blobs:     make([]blob, 0),
			pieces:    make([]piece, 0),
		},
		mutex: &sync.Mutex{},
	}
}

// Register implements gophkeeper.Gophkeeper.
func (k *Gophkeeper) Register(_ context.Context, credential gophkeeper.Credential) error {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	identityID := k.findIdentity(credential.Username)
	if identityID != invalidIdentityID {
		return gophkeeper.ErrIdentityDuplicate
	}

	k.identities = append(
		k.identities,
		identity{
			username: credential.Username,
			password: credential.Password,
		},
	)
	return nil
}

// Authenticate implements gophkeeper.Gophkeeper.
func (k *Gophkeeper) Authenticate(ctx context.Context, credential gophkeeper.Credential) (gophkeeper.Token, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	id := k.findIdentity(credential.Username)
	if id == invalidIdentityID {
		return gophkeeper.InvalidToken, gophkeeper.ErrBadCredential
	}

	return k.ts.Create(ctx, credential.Username)
}

// Identity implements gophkeeper.Gophkeeper.
func (k *Gophkeeper) Identity(ctx context.Context, token gophkeeper.Token) (gophkeeper.Identity, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	username, usernameError := k.ts.Unwrap(ctx, token)
	if usernameError != nil {
		return nil, errors.Join(usernameError, gophkeeper.ErrBadCredential)
	}

	identity := &Identity{
		identity: k.identities[k.findIdentity(username)],
		storage:  k.storage,
		blobsDir: k.blobsDir,
	}
	return identity, nil
}

func (k *Gophkeeper) findIdentity(username string) int {
	for i := range k.identities {
		identity := k.identities[i]
		if identity.username == username {
			return i
		}
	}
	return invalidIdentityID
}
