// Package register provides REST endpoint to register
// a new user.
package register

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// Entry is login entry.
type Entry struct {
	Gophkeeper gophkeeper.Gophkeeper
}

// Route routes login entry.
func (e *Entry) Route() http.Handler {
	router := chi.NewRouter()
	router.Post("/", e.register)
	return router
}

func (e *Entry) register(out http.ResponseWriter, in *http.Request) {
	var requestBody struct {
		Username *string `json:"username"`
		Password *string `json:"password"`
	}
	if err := json.NewDecoder(in.Body).Decode(&requestBody); err != nil {
		status := http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}

	var credential gophkeeper.Credential
	if username := requestBody.Username; username != nil {
		credential.Username = *username
	} else {
		status := http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}

	if password := requestBody.Password; password != nil {
		credential.Password = *password
	} else {
		status := http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}

	if err := e.Gophkeeper.Register(in.Context(), credential); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gophkeeper.ErrBadCredential) {
			status = http.StatusBadRequest
		}
		if errors.Is(err, gophkeeper.ErrIdentityDuplicate) {
			status = http.StatusConflict
		}
		http.Error(out, http.StatusText(status), status)
		return
	}

	out.WriteHeader(http.StatusCreated)
}
