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
	var router = chi.NewRouter()
	router.Post("/", e.post)
	return router
}

func (e *Entry) post(out http.ResponseWriter, in *http.Request) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(in.Body).Decode(&request); err != nil {
		var status = http.StatusBadRequest
		http.Error(out, http.StatusText(status), status)
		return
	}

	var credential = gophkeeper.Credential{
		Username: request.Username,
		Password: request.Password,
	}
	if err := e.Gophkeeper.Register(in.Context(), credential); err != nil {
		var status = http.StatusInternalServerError
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
