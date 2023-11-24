package rest_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kerelape/gophkeeper/internal/server/rest"
	"github.com/kerelape/gophkeeper/pkg/gophkeeper/virtual"
	"github.com/stretchr/testify/assert"
)

func TestEntry(t *testing.T) {
	r := rest.Entry{
		Gophkeeper: virtual.New(time.Hour, t.TempDir()),
	}
	handler := r.Route()
	assert.NotNil(t, handler, "unexpected nil")

	var token string

	t.Run("register", func(t *testing.T) {
		register := func() *http.Response {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(
					http.MethodPost,
					"/register",
					strings.NewReader(`{"username": "test", "password": "qwerty"}`),
				)
			)
			handler.ServeHTTP(recorder, request)
			return recorder.Result()
		}
		t.Run("New identity", func(t *testing.T) {
			response := register()
			assert.Equal(t, http.StatusCreated, response.StatusCode)
		})
		t.Run("Duplicate identity", func(t *testing.T) {
			response := register()
			assert.Equal(t, http.StatusConflict, response.StatusCode)
		})
		t.Run("Non-JSON body", func(t *testing.T) {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(
					http.MethodPost,
					"/register",
					strings.NewReader(`_`),
				)
			)
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()
			assert.Equal(t, http.StatusBadRequest, response.StatusCode, "unexpected status code")
		})
		t.Run("Without password", func(t *testing.T) {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(
					http.MethodPost,
					"/register",
					strings.NewReader(`{"username":"test"}`),
				)
			)
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()
			assert.Equal(t, http.StatusBadRequest, response.StatusCode, "unexpected status code")
		})
		t.Run("Without username", func(t *testing.T) {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(
					http.MethodPost,
					"/register",
					strings.NewReader(`{"password":"qwerty"}`),
				)
			)
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()
			assert.Equal(t, http.StatusBadRequest, response.StatusCode, "unexpected status code")
		})
	})
	t.Run("login", func(t *testing.T) {
		t.Run("Existing identity", func(t *testing.T) {
			t.Run("Non-JSON body", func(t *testing.T) {
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(
						http.MethodPost,
						"/login",
						strings.NewReader(`_`),
					)
				)
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusBadRequest, response.StatusCode, "expected to successfully login")
			})
			t.Run("Invalid body", func(t *testing.T) {
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(
						http.MethodPost,
						"/login",
						strings.NewReader(`{}`),
					)
				)
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusBadRequest, response.StatusCode, "expected to successfully login")
			})
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(
					http.MethodPost,
					"/login",
					strings.NewReader(`{"username": "test", "password": "qwerty"}`),
				)
			)
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()

			assert.Equal(t, http.StatusOK, response.StatusCode, "expected to successfully login")
			assert.NotEmpty(t, response.Header.Get("Authorization"), "missing token")
			token = response.Header.Get("Authorization")
		})
		t.Run("Unexisting identity", func(t *testing.T) {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(
					http.MethodPost,
					"/login",
					strings.NewReader(`{"username": "alian", "password": "qwerty"}`),
				)
			)
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()

			assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "expected NOT to login")
		})
	})
	t.Run("vault", func(t *testing.T) {
		t.Run("piece", func(t *testing.T) {
			var rid int
			t.Run("Store", func(t *testing.T) {
				t.Run("Non-JSON body", func(t *testing.T) {
					var (
						recorder = httptest.NewRecorder()
						request  = httptest.NewRequest(
							http.MethodPut,
							"/vault/piece",
							strings.NewReader("_"),
						)
					)
					request.Header.Set("Authorization", token)
					request.Header.Set("X-Password", "qwerty")
					handler.ServeHTTP(recorder, request)
					response := recorder.Result()
					assert.Equal(t, http.StatusBadRequest, response.StatusCode, "unexpected status code")
				})
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(
						http.MethodPut,
						"/vault/piece",
						strings.NewReader(
							fmt.Sprintf(
								`{"meta": "testmeta", "content": "%s"}`,
								base64.RawStdEncoding.EncodeToString(([]byte)("Hello, World!")),
							),
						),
					)
				)
				request.Header.Set("Authorization", token)
				request.Header.Set("X-Password", "qwerty")
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusCreated, response.StatusCode, "unexpected status code")

				var responseBody struct {
					RID int `json:"rid"`
				}
				decodeError := json.NewDecoder(response.Body).Decode(&responseBody)
				assert.Nil(t, decodeError, "did not expect an error")
				rid = responseBody.RID
			})
			t.Run("Restore", func(t *testing.T) {
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(
						http.MethodGet,
						fmt.Sprintf("/vault/piece/%d", rid),
						nil,
					)
				)
				request.Header.Set("Authorization", token)
				request.Header.Set("X-Password", "qwerty")
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusOK, response.StatusCode, "unexpected status code")

				var responseBody struct {
					Meta    string `json:"meta"`
					Content string `json:"content"`
				}
				decodeError := json.NewDecoder(response.Body).Decode(&responseBody)
				assert.Nil(t, decodeError, "did not expect an error")

				content, contentError := base64.RawStdEncoding.DecodeString(responseBody.Content)
				assert.Nil(t, contentError, "did not expect an error")

				assert.Equal(t, "testmeta", responseBody.Meta, "meta is not correct")
				assert.Equal(t, "Hello, World!", (string)(content), "content is not correct")
			})
		})

		t.Run("blob", func(t *testing.T) {
			var rid int
			t.Run("Store", func(t *testing.T) {
				t.Run("Without token", func(t *testing.T) {
					var (
						recorder = httptest.NewRecorder()
						request  = httptest.NewRequest(
							http.MethodPut,
							"/vault/blob",
							strings.NewReader("Hello, World!"),
						)
					)
					handler.ServeHTTP(recorder, request)
					response := recorder.Result()
					assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected status code")
				})
				t.Run("Without password", func(t *testing.T) {
					var (
						recorder = httptest.NewRecorder()
						request  = httptest.NewRequest(
							http.MethodPut,
							"/vault/blob",
							strings.NewReader("Hello, World!"),
						)
					)
					request.Header.Set("X-Password", "qwerty")
					handler.ServeHTTP(recorder, request)
					response := recorder.Result()
					assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected status code")
				})
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(
						http.MethodPut,
						"/vault/blob",
						strings.NewReader("Hello, World!"),
					)
				)
				request.Header.Set("Authorization", token)
				request.Header.Set("X-Password", "qwerty")
				request.Header.Set("X-Meta", "testmeta")
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusCreated, response.StatusCode, "unexpected status code")
				var responseBody struct {
					RID int `json:"rid"`
				}
				decodeError := json.NewDecoder(response.Body).Decode(&responseBody)
				assert.Nil(t, decodeError, "did not expect an error")
				rid = responseBody.RID
			})
			t.Run("Restore", func(t *testing.T) {
				t.Run("Without token", func(t *testing.T) {
					var (
						recorder = httptest.NewRecorder()
						request  = httptest.NewRequest(
							http.MethodGet,
							fmt.Sprintf("/vault/blob/%d", rid),
							nil,
						)
					)
					handler.ServeHTTP(recorder, request)
					response := recorder.Result()
					assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected status code")
				})
				t.Run("Without password", func(t *testing.T) {
					var (
						recorder = httptest.NewRecorder()
						request  = httptest.NewRequest(
							http.MethodGet,
							fmt.Sprintf("/vault/blob/%d", rid),
							nil,
						)
					)
					request.Header.Set("Authorization", token)
					handler.ServeHTTP(recorder, request)
					response := recorder.Result()
					assert.Equal(t, http.StatusBadRequest, response.StatusCode, "unexpected status code")
				})
				t.Run("Invalid RID", func(t *testing.T) {
					var (
						recorder = httptest.NewRecorder()
						request  = httptest.NewRequest(http.MethodGet, "/vault/blob/_", nil)
					)
					request.Header.Set("Authorization", token)
					request.Header.Set("X-Password", "qwerty")
					handler.ServeHTTP(recorder, request)
					response := recorder.Result()
					assert.Equal(t, http.StatusBadRequest, response.StatusCode, "unexpected status code")
				})
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(
						http.MethodGet,
						fmt.Sprintf("/vault/blob/%d", rid),
						nil,
					)
				)
				request.Header.Set("Authorization", token)
				request.Header.Set("X-Password", "qwerty")
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusOK, response.StatusCode, "unexpected status code")

				assert.Equal(t, "testmeta", response.Header.Get("X-Meta"), "meta is not correct")

				content, contentError := io.ReadAll(response.Body)
				assert.Nil(t, contentError, "unexpected error")
				assert.Equal(t, "Hello, World!", (string)(content), "content is not correct")
			})
		})

		t.Run("List", func(t *testing.T) {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(http.MethodGet, "/vault", nil)
			)
			request.Header.Set("Authorization", token)
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()
			assert.Equal(t, http.StatusOK, response.StatusCode, "unexpected status code")
			t.Run("Without token", func(t *testing.T) {
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(http.MethodGet, "/vault", nil)
				)
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected status code")
			})
		})

		t.Run("Delete", func(t *testing.T) {
			var (
				recorder = httptest.NewRecorder()
				request  = httptest.NewRequest(http.MethodDelete, "/vault/0", nil)
			)
			request.Header.Set("Authorization", token)
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()
			assert.Equal(t, http.StatusOK, response.StatusCode, "unexpected status code")
			t.Run("Without token", func(t *testing.T) {
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(http.MethodDelete, "/vault/0", nil)
				)
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusUnauthorized, response.StatusCode, "unexpected status code")
			})
			t.Run("Invalid RID", func(t *testing.T) {
				var (
					recorder = httptest.NewRecorder()
					request  = httptest.NewRequest(http.MethodDelete, "/vault/_", nil)
				)
				request.Header.Set("Authorization", token)
				handler.ServeHTTP(recorder, request)
				response := recorder.Result()
				assert.Equal(t, http.StatusBadRequest, response.StatusCode, "unexpected status code")
			})
		})
	})
}
