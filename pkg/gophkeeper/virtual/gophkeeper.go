package virtual

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

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
	sessions   map[gophkeeper.Token]int

	sessionLifespan time.Duration
	blobsDir        string

	storage *storage

	mutex *sync.Mutex
}

// New returns a new virtual Gophkeeper that store all its
// data in RAM.
func New(sessionLifespan time.Duration, blobsDir string) Gophkeeper {
	return Gophkeeper{
		identities:      make([]identity, 0),
		sessions:        make(map[gophkeeper.Token]int),
		sessionLifespan: sessionLifespan,
		blobsDir:        blobsDir,
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

	var identityID = k.findIdentity(credential.Username)
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
func (k *Gophkeeper) Authenticate(_ context.Context, credential gophkeeper.Credential) (gophkeeper.Token, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	var id = k.findIdentity(credential.Username)
	if id == invalidIdentityID {
		return (gophkeeper.Token)(""), gophkeeper.ErrBadCredential
	}

	var sessionBytes = make([]byte, 2048)
	if _, err := rand.Read(sessionBytes); err != nil {
		return (gophkeeper.Token)(""), err
	}

	var session = (gophkeeper.Token)(base64.URLEncoding.EncodeToString(sessionBytes))
	k.sessions[session] = id

	go func(sessions map[gophkeeper.Token]int, session gophkeeper.Token) {
		time.Sleep(k.sessionLifespan)
		k.mutex.Lock()
		delete(k.sessions, session)
		k.mutex.Unlock()
	}(k.sessions, session)

	return session, nil
}

// Identity implements gophkeeper.Gophkeeper.
func (k *Gophkeeper) Identity(_ context.Context, token gophkeeper.Token) (gophkeeper.Identity, error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	if _, ok := k.sessions[token]; !ok {
		return nil, gophkeeper.ErrBadCredential
	}
	var identity = &Identity{
		identity: k.identities[k.sessions[token]],
		storage:  k.storage,
		blobsDir: k.blobsDir,
	}
	return identity, nil
}

func (k *Gophkeeper) findIdentity(username string) int {
	for i := range k.identities {
		var identity = k.identities[i]
		if identity.username == username {
			return i
		}
	}
	return invalidIdentityID
}
