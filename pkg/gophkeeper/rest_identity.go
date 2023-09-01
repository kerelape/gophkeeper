package gophkeeper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// RestIdentity is rest identity.
type RestIdentity struct {
	Client http.Client
	Server string
	Token  Token
}

var _ Identity = (*RestIdentity)(nil)

// StorePiece implements Identity.
func (i *RestIdentity) StorePiece(ctx context.Context, piece Piece, password string) (ResourceID, error) {
	var endpoint = fmt.Sprintf("%s/piece", i.Server)
	var content, contentError = json.Marshal(
		map[string]any{
			"meta":     piece.Meta,
			"content":  base64.RawStdEncoding.EncodeToString(([]byte)(piece.Content)),
			"password": password,
		},
	)
	if contentError != nil {
		return -1, contentError
	}
	var request, requestError = http.NewRequestWithContext(
		ctx,
		http.MethodPut, endpoint,
		bytes.NewReader(content),
	)
	if requestError != nil {
		return -1, requestError
	}
	request.Header.Set("Authorization", (string)(i.Token))
	var response, responseError = i.Client.Do(request)
	if responseError != nil {
		return -1, responseError
	}
	switch response.StatusCode {
	case http.StatusCreated:
		var content = make(map[string]any)
		if err := json.NewDecoder(response.Body).Decode(&content); err != nil {
			return -1, errors.Join(
				fmt.Errorf("parse response: %w", err),
				ErrIncompatibleAPI,
			)
		}
		if rid, ok := content["rid"].(int64); ok {
			return (ResourceID)(rid), nil
		}
		return -1, errors.Join(
			fmt.Errorf("invalid response"),
			ErrIncompatibleAPI,
		)
	case http.StatusUnauthorized:
		return -1, ErrBadCredential
	default:
		return -1, errors.Join(
			fmt.Errorf("unexpected response status: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// RestorePiece implements Identity.
func (i *RestIdentity) RestorePiece(ctx context.Context, rid ResourceID, password string) (Piece, error) {
	var endpoint = fmt.Sprintf("%s/piece/%d", i.Server, rid)
	var content, contentError = json.Marshal(
		map[string]any{
			"password": password,
		},
	)
	if contentError != nil {
		return Piece{}, contentError
	}
	var request, requestError = http.NewRequestWithContext(
		ctx,
		http.MethodGet, endpoint,
		bytes.NewReader(content),
	)
	if requestError != nil {
		return Piece{}, requestError
	}
	request.Header.Set("Authorization", (string)(i.Token))
	var response, responseError = i.Client.Do(request)
	if responseError != nil {
		return Piece{}, responseError
	}
	switch response.StatusCode {
	case http.StatusCreated:
		var content = make(map[string]any)
		if err := json.NewDecoder(response.Body).Decode(&content); err != nil {
			return Piece{}, errors.Join(
				fmt.Errorf("parse response: %w", err),
				ErrIncompatibleAPI,
			)
		}
		var piece Piece
		if meta, ok := content["meta"].(string); ok {
			piece.Meta = meta
		} else {
			return Piece{}, errors.Join(
				fmt.Errorf("invalid response body"),
				ErrIncompatibleAPI,
			)
		}
		if content, ok := content["content"].(string); ok {
			var decodedContent, decodedContentError = base64.RawStdEncoding.DecodeString(content)
			if decodedContentError != nil {
				return Piece{}, errors.Join(
					fmt.Errorf("decode content: %w", decodedContentError),
					ErrIncompatibleAPI,
				)
			}
			piece.Content = decodedContent
		} else {
			return Piece{}, errors.Join(
				fmt.Errorf("invalid response body"),
				ErrIncompatibleAPI,
			)
		}
		return piece, nil
	case http.StatusUnauthorized:
		return Piece{}, ErrBadCredential
	default:
		return Piece{}, errors.Join(
			fmt.Errorf("unexpected response status: %d", response.StatusCode),
			ErrIncompatibleAPI,
		)
	}
}

// StoreBlob implements Identity.
//
// @todo #31 Implement StoreBlob on RestIdentity.
func (i *RestIdentity) StoreBlob(ctx context.Context, blob Blob, password string) (ResourceID, error) {
	panic("unimplemented")
}

// RestoreBlob implements Identity.
//
// @todo #31 Implement RestoreBlob on RestIdentity.
func (i *RestIdentity) RestoreBlob(ctx context.Context, rid ResourceID, password string) (Blob, error) {
	panic("unimplemented")
}

// Delete implements Identity.
//
// @todo #31 Implement Delete on RestIdentity.
func (i *RestIdentity) Delete(context.Context, ResourceID) error {
	panic("unimplemented")
}

// List implements Identity.
//
// @todo #31 Implement List on RestIdentity.
func (i *RestIdentity) List(context.Context) ([]Resource, error) {
	panic("unimplemented")
}
