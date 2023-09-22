package piece

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kerelape/gophkeeper/internal/server/rest/authentication"
	"github.com/kerelape/gophkeeper/internal/server/rest/vault/credential"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// Entry is piece entry.
type Entry struct{}

// Route routes piece entry.
func (e *Entry) Route() http.Handler {
	var router = chi.NewRouter()
	router.Use(credential.Middleware)
	router.Put("/", e.encrypt)
	router.Get("/{rid}", e.decrypt)
	return router
}

func (e *Entry) encrypt(out http.ResponseWriter, in *http.Request) {
	identity := authentication.Identity(in)
	password := credential.Password(in)

	var request struct {
		Meta    string `json:"meta"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(in.Body).Decode(&request); err != nil {
		var status = http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}
	var content = make([]byte, len(request.Content))
	if _, err := base64.RawStdEncoding.Decode(content, ([]byte)(request.Content)); err != nil {
		var status = http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}

	var piece = gophkeeper.Piece{
		Meta:    request.Meta,
		Content: content,
	}

	var rid, storeError = identity.StorePiece(in.Context(), piece, password)
	if storeError != nil {
		var status = http.StatusInternalServerError
		if errors.Is(storeError, gophkeeper.ErrBadCredential) {
			status = http.StatusUnauthorized
		}
		http.Error(out, http.StatusText(status), status)
		return
	}

	out.WriteHeader(http.StatusCreated)
	var response struct {
		RID int64 `json:"rid"`
	}
	response.RID = (int64)(rid)
	if err := json.NewEncoder(out).Encode(&response); err != nil {
		log.Printf("Failed to write response: %s", err.Error())
	}
}

func (e *Entry) decrypt(out http.ResponseWriter, in *http.Request) {
	identity := authentication.Identity(in)
	password := credential.Password(in)

	var rid, ridError = strconv.Atoi(chi.URLParam(in, "rid"))
	if ridError != nil {
		var status = http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}

	var piece, restoreError = identity.RestorePiece(in.Context(), (gophkeeper.ResourceID)(rid), password)
	if restoreError != nil {
		var status = http.StatusInternalServerError
		if errors.Is(restoreError, gophkeeper.ErrBadCredential) {
			status = http.StatusUnauthorized
		}
		if errors.Is(restoreError, gophkeeper.ErrResourceNotFound) {
			status = http.StatusNotFound
		}
		http.Error(out, http.StatusText(status), status)
		return
	}

	var response struct {
		Meta    string `json:"meta"`
		Content string `json:"content"`
	}
	response.Meta = piece.Meta
	response.Content = base64.RawStdEncoding.EncodeToString(
		bytes.ReplaceAll(
			piece.Content,
			[]byte{'\x00'},
			[]byte{},
		),
	)
	out.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(out).Encode(response); err != nil {
		log.Printf("Failed to write response: %s", err.Error())
	}
}
