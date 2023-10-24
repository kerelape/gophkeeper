package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// ErrIncompatibleAPI is returns when API is not compatible with implementation.
var ErrIncompatibleAPI = errors.New("incompatable API")

// Gophkeeper is a remote gophkeeper.
type Gophkeeper struct {
	Client http.Client
	Server string
}

var _ gophkeeper.Gophkeeper = (*Gophkeeper)(nil)

// Register implements Gophkeeper.
func (g *Gophkeeper) Register(ctx context.Context, credential gophkeeper.Credential) error {
	endpoint := fmt.Sprintf("%s/register", g.Server)
	content, marshalError := json.Marshal(
		map[string]any{
			"username": credential.Username,
			"password": credential.Password,
		},
	)
	if marshalError != nil {
		return marshalError
	}
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodPost, endpoint,
		bytes.NewReader(content),
	)
	if requestError != nil {
		return requestError
	}
	response, postError := g.Client.Do(request)
	if postError != nil {
		return postError
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusConflict:
		return gophkeeper.ErrIdentityDuplicate
	case http.StatusCreated:
		return nil
	default:
		return errors.Join(
			fmt.Errorf("unexpected response status: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// Authenticate implements Gophkeeper.
func (g *Gophkeeper) Authenticate(ctx context.Context, credential gophkeeper.Credential) (gophkeeper.Token, error) {
	endpoint := fmt.Sprintf("%s/login", g.Server)
	content, marshalError := json.Marshal(
		map[string]any{
			"username": credential.Username,
			"password": credential.Password,
		},
	)
	if marshalError != nil {
		return (gophkeeper.Token)(""), marshalError
	}
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodPost, endpoint,
		bytes.NewReader(content),
	)
	if requestError != nil {
		return (gophkeeper.Token)(""), requestError
	}
	response, postError := g.Client.Do(request)
	if postError != nil {
		return (gophkeeper.Token)(""), postError
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusUnauthorized:
		return (gophkeeper.Token)(""), gophkeeper.ErrBadCredential
	case http.StatusOK:
		token := response.Header.Get("Authorization")
		return (gophkeeper.Token)(token), nil
	default:
		return (gophkeeper.Token)(""), errors.Join(
			fmt.Errorf("unexpected response status: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// Identity implements Gophkeeper.
func (g *Gophkeeper) Identity(_ context.Context, token gophkeeper.Token) (gophkeeper.Identity, error) {
	identity := &Identity{
		Client: g.Client,
		Server: g.Server,
		Token:  token,
	}
	return identity, nil
}
