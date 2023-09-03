package vault

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kerelape/gophkeeper/internal/server/rest/vault/blob"
	"github.com/kerelape/gophkeeper/internal/server/rest/vault/piece"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper"
)

// Entry is vault entry.
type Entry struct {
	Gophkeeper gophkeeper.Gophkeeper
}

// Route routes vault entry.
func (e *Entry) Route() http.Handler {
	var (
		piece = piece.Entry{
			Gophkeeper: e.Gophkeeper,
		}
		blob = blob.Entry{
			Gophkeeper: e.Gophkeeper,
		}
	)
	var router = chi.NewRouter()
	router.Mount("/piece", piece.Route())
	router.Mount("/blob", blob.Route())
	router.Get("/", nil)
	router.Delete("/{rid}", nil)
	return router
}

func (e *Entry) get(out http.ResponseWriter, in *http.Request) {
	var token = in.Header.Get("Authorization")
	var identity, identityError = e.Gophkeeper.Identity(in.Context(), (gophkeeper.Token)(token))
	if identityError != nil {
		var status = http.StatusInternalServerError
		if errors.Is(identityError, gophkeeper.ErrBadCredential) {
			status = http.StatusUnauthorized
		}
		http.Error(out, http.StatusText(status), status)
		return
	}

	var resources, resourcesError = identity.List(in.Context())
	if resourcesError != nil {
		var status = http.StatusInternalServerError
		http.Error(out, http.StatusText(status), status)
		return
	}

	var response = make([](map[string]any), len(resources))
	for _, resource := range resources {
		response = append(
			response,
			map[string]any{
				"rid":  (int64)(resource.ID),
				"meta": resource.Meta,
				"type": (int)(resource.Type),
			},
		)
	}

	out.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(out).Encode(&response); err != nil {
		log.Printf("failed to write response: %s\n", err.Error())
	}
}

// @todo #32 Implement `DELETE /vault/{rid}`
func (e *Entry) delete(out http.ResponseWriter, in *http.Request) {
	panic("unimplemented")
}
