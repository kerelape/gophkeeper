package piece

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/kerelape/gophkeeper/internal/server/domain"
)

// Entry is piece entry.
type Entry struct {
	Repository domain.Repository
}

// Route routes piece entry.
func (e *Entry) Route() http.Handler {
	var router = chi.NewRouter()
	router.Put("/", e.encrypt)
	router.Get("/{rid}", e.decrypt)
	return router
}

func (e *Entry) encrypt(out http.ResponseWriter, in *http.Request) {
	var token = in.Header.Get("Authorization")
	var identity, identityError = e.Repository.Identity(in.Context(), (domain.Token)(token))
	if identityError != nil {
		var status = http.StatusInternalServerError
		if errors.Is(identityError, domain.ErrBadCredential) {
			status = http.StatusUnauthorized
		}
		http.Error(out, http.StatusText(status), status)
		return
	}

	var request struct {
		Meta     string `json:"meta"`
		Content  string `json:"content"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(in.Body).Decode(&request); err != nil {
		var status = http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}
	var content []byte
	if _, err := base64.RawStdEncoding.Decode(content, ([]byte)(request.Content)); err != nil {
		var status = http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}

	var piece = domain.Piece{
		Meta:    request.Meta,
		Content: content,
	}
	var rid, storeError = identity.StorePiece(in.Context(), piece, request.Password)
	if storeError != nil {
		var status = http.StatusInternalServerError
		if errors.Is(identityError, domain.ErrBadCredential) {
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
	if err := json.NewEncoder(out).Encode(response); err != nil {
		log.Printf("Failed to write response: %s", err.Error())
	}
}

func (e *Entry) decrypt(out http.ResponseWriter, in *http.Request) {
	var token = in.Header.Get("Authorization")
	var identity, identityError = e.Repository.Identity(in.Context(), (domain.Token)(token))
	if identityError != nil {
		var status = http.StatusInternalServerError
		if errors.Is(identityError, domain.ErrBadCredential) {
			status = http.StatusUnauthorized
		}
		http.Error(out, http.StatusText(status), status)
		return
	}

	var request struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(in.Body).Decode(&request); err != nil {
		var status = http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}
	var rid, ridError = strconv.Atoi(chi.URLParam(in, "rid"))
	if ridError != nil {
		var status = http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}

	var piece, restoreError = identity.RestorePiece(in.Context(), (domain.ResourceID)(rid), request.Password)
	if restoreError != nil {
		var status = http.StatusInternalServerError
		if errors.Is(identityError, domain.ErrBadCredential) {
			status = http.StatusUnauthorized
		}
		http.Error(out, http.StatusText(status), status)
		return
	}

	var response struct {
		Meta    string `json:"meta"`
		Content string `json:"content"`
	}
	response.Meta = piece.Meta
	response.Content = base64.RawStdEncoding.EncodeToString(piece.Content)
	out.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(out).Encode(response); err != nil {
		log.Printf("Failed to write response: %s", err.Error())
	}
}
