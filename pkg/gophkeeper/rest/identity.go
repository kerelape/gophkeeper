package rest

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// ErrServerIsDown is returns when server returned an internal server error.
var ErrServerIsDown = errors.New("server is down")

// Identity is rest identity.
type Identity struct {
	Client http.Client
	Server string
	Token  gophkeeper.Token
}

var _ gophkeeper.Identity = (*Identity)(nil)

// StorePiece implements Identity.
func (i *Identity) StorePiece(ctx context.Context, piece gophkeeper.Piece, password string) (gophkeeper.ResourceID, error) {
	endpoint := fmt.Sprintf("%s/vault/piece", i.Server)
	content, contentError := json.Marshal(
		map[string]any{
			"meta":    piece.Meta,
			"content": base64.RawStdEncoding.EncodeToString(piece.Content),
		},
	)
	if contentError != nil {
		return -1, contentError
	}
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodPut, endpoint,
		bytes.NewReader(content),
	)
	if requestError != nil {
		return -1, requestError
	}
	request.Header.Set("Authorization", (string)(i.Token))
	request.Header.Set("X-Password", password)

	response, responseError := i.Client.Do(request)
	if responseError != nil {
		return -1, responseError
	}
	switch response.StatusCode {
	case http.StatusCreated:
		var content struct {
			RID gophkeeper.ResourceID `json:"rid"`
		}
		if err := json.NewDecoder(response.Body).Decode(&content); err != nil {
			return -1, errors.Join(
				fmt.Errorf("parse response: %w", err),
				ErrIncompatibleAPI,
			)
		}
		return content.RID, nil
	case http.StatusUnauthorized:
		return -1, gophkeeper.ErrBadCredential
	default:
		return -1, errors.Join(
			fmt.Errorf("unexpected response status: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// RestorePiece implements Identity.
func (i *Identity) RestorePiece(ctx context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Piece, error) {
	endpoint := fmt.Sprintf("%s/vault/piece/%d", i.Server, rid)
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodGet, endpoint,
		nil,
	)
	if requestError != nil {
		return gophkeeper.Piece{}, requestError
	}
	request.Header.Set("Authorization", (string)(i.Token))
	request.Header.Set("X-Password", password)

	response, responseError := i.Client.Do(request)
	if responseError != nil {
		return gophkeeper.Piece{}, responseError
	}
	switch response.StatusCode {
	case http.StatusOK:
		content := make(map[string]any)
		if err := json.NewDecoder(response.Body).Decode(&content); err != nil {
			return gophkeeper.Piece{}, errors.Join(
				fmt.Errorf("parse response: %w", err),
				ErrIncompatibleAPI,
			)
		}
		var piece gophkeeper.Piece
		if meta, ok := content["meta"].(string); ok {
			piece.Meta = meta
		} else {
			return gophkeeper.Piece{}, errors.Join(
				fmt.Errorf("invalid response body"),
				ErrIncompatibleAPI,
			)
		}
		if content, ok := content["content"].(string); ok {
			decodedContent, decodedContentError := base64.RawStdEncoding.DecodeString(content)
			if decodedContentError != nil {
				return gophkeeper.Piece{}, errors.Join(
					fmt.Errorf("decode content: %w", decodedContentError),
					ErrIncompatibleAPI,
				)
			}
			piece.Content = decodedContent
		} else {
			return gophkeeper.Piece{}, errors.Join(
				fmt.Errorf("invalid response body"),
				ErrIncompatibleAPI,
			)
		}
		return piece, nil
	case http.StatusUnauthorized:
		return gophkeeper.Piece{}, gophkeeper.ErrBadCredential
	case http.StatusNotFound:
		return gophkeeper.Piece{}, gophkeeper.ErrResourceNotFound
	default:
		return gophkeeper.Piece{}, errors.Join(
			fmt.Errorf("unexpected response status: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// StoreBlob implements Identity.
func (i *Identity) StoreBlob(ctx context.Context, blob gophkeeper.Blob, password string) (gophkeeper.ResourceID, error) {
	endpoint := fmt.Sprintf("%s/vault/blob", i.Server)
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodPut, endpoint,
		blob.Content,
	)
	if requestError != nil {
		return -1, requestError
	}
	request.Header.Set("Authorization", (string)(i.Token))
	request.Header.Set("X-Password", password)
	request.Header.Set("X-Meta", blob.Meta)

	response, responseError := i.Client.Do(request)
	if responseError != nil {
		return -1, responseError
	}

	switch response.StatusCode {
	case http.StatusCreated:
		var content struct {
			RID gophkeeper.ResourceID `json:"rid"`
		}
		if err := json.NewDecoder(response.Body).Decode(&content); err != nil {
			return -1, ErrIncompatibleAPI
		}
		return content.RID, nil
	case http.StatusUnauthorized:
		return -1, gophkeeper.ErrBadCredential
	default:
		return -1, errors.Join(
			fmt.Errorf("unexpected response status: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// RestoreBlob implements Identity.
func (i *Identity) RestoreBlob(ctx context.Context, rid gophkeeper.ResourceID, password string) (gophkeeper.Blob, error) {
	endpoint := fmt.Sprintf("%s/vault/blob/%d", i.Server, rid)
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodGet, endpoint,
		nil,
	)
	if requestError != nil {
		return gophkeeper.Blob{}, requestError
	}
	request.Header.Set("Authorization", (string)(i.Token))
	request.Header.Set("X-Password", password)

	response, responseError := i.Client.Do(request)
	if responseError != nil {
		return gophkeeper.Blob{}, responseError
	}
	switch response.StatusCode {
	case http.StatusOK:
		blob := gophkeeper.Blob{
			Meta:    response.Header.Get("X-Meta"),
			Content: response.Body,
		}
		return blob, nil
	case http.StatusUnauthorized:
		return gophkeeper.Blob{}, gophkeeper.ErrBadCredential
	case http.StatusNotFound:
		return gophkeeper.Blob{}, gophkeeper.ErrResourceNotFound
	default:
		return gophkeeper.Blob{}, errors.Join(
			fmt.Errorf("unexpected response status: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// Delete implements Identity.
func (i *Identity) Delete(ctx context.Context, rid gophkeeper.ResourceID) error {
	endpoint := fmt.Sprintf("%s/vault/%d", i.Server, rid)
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodDelete, endpoint,
		nil,
	)
	if requestError != nil {
		return requestError
	}
	request.Header.Set("Authorization", (string)(i.Token))

	response, responseError := i.Client.Do(request)
	if responseError != nil {
		return responseError
	}

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return gophkeeper.ErrResourceNotFound
	default:
		return errors.Join(
			fmt.Errorf("unexpected response code: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// List implements Identity.
func (i *Identity) List(ctx context.Context) ([]gophkeeper.Resource, error) {
	endpoint := fmt.Sprintf("%s/vault", i.Server)
	request, requestError := http.NewRequestWithContext(
		ctx,
		http.MethodGet, endpoint,
		nil,
	)
	if requestError != nil {
		return nil, requestError
	}
	request.Header.Set("Authorization", (string)(i.Token))

	response, responseError := i.Client.Do(request)
	if responseError != nil {
		return nil, responseError
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		responseContent := make(
			[]struct {
				Meta string                  `json:"meta"`
				RID  gophkeeper.ResourceID   `json:"rid"`
				Type gophkeeper.ResourceType `json:"type"`
			},
			0,
		)
		if err := json.NewDecoder(response.Body).Decode(&responseContent); err != nil {
			return nil, err
		}
		resources := make([]gophkeeper.Resource, 0, len(responseContent))
		for _, responseResource := range responseContent {
			resources = append(
				resources,
				gophkeeper.Resource{
					ID:   responseResource.RID,
					Type: responseResource.Type,
					Meta: responseResource.Meta,
				},
			)
		}
		return resources, nil
	default:
		return nil, errors.Join(
			fmt.Errorf("unexpected response code: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}
