package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kerelape/gophkeeper/internal/server/domain"
	"github.com/kerelape/gophkeeper/internal/server/rest/login"
	"github.com/kerelape/gophkeeper/internal/server/rest/register"
)

// Entry is the REST api entry.
//
// @todo #7 Implement text storage.
// @todo #7 Implement blob storage.
// @todo #7 Implement bank card storage.
type Entry struct {
	Repository domain.Repository
}

// Route routes Entry into an http.Handler.
func (e *Entry) Route() http.Handler {
	var (
		registretion = register.Entry{
			Repository: e.Repository,
		}
		authentication = login.Entry{
			Repository: e.Repository,
		}
	)
	var router = chi.NewRouter()
	router.Mount("/register", registretion.Route())
	router.Mount("/login", authentication.Route())
	return router
}
