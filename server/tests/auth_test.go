package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth_Register(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("201 – valid registration", func(t *testing.T) {
		body := map[string]string{
			"email":    "alice@example.com",
			"password": "secret123",
			"name":     "Alice",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/auth/register", body, "")
		assertStatus(t, w, http.StatusCreated)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "registered")
	})

	t.Run("400 – missing required fields", func(t *testing.T) {
		body := map[string]string{"email": "bob@example.com"}
		w := doRequest(env, http.MethodPost, "/api/v1/auth/register", body, "")
		assertStatus(t, w, http.StatusBadRequest)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.False(t, resp.Success)
	})

	t.Run("400 – invalid email format", func(t *testing.T) {
		body := map[string]string{
			"email":    "not-an-email",
			"password": "secret123",
			"name":     "Bob",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/auth/register", body, "")
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("409 – duplicate email", func(t *testing.T) {
		body := map[string]string{
			"email":    "duplicate@example.com",
			"password": "secret123",
			"name":     "Dup",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/auth/register", body, "")
		assertStatus(t, w, http.StatusCreated)

		w2 := doRequest(env, http.MethodPost, "/api/v1/auth/register", body, "")
		assertStatus(t, w2, http.StatusConflict)
	})
}

func TestAuth_Login(t *testing.T) {
	env := setupTestEnv(t)
	seedUser(t, env, "login@example.com", "pass1234", "LoginUser")

	t.Run("200 – valid credentials", func(t *testing.T) {
		body := map[string]string{
			"email":    "login@example.com",
			"password": "pass1234",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/auth/login", body, "")
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "Login")
	})

	t.Run("400 – missing password field", func(t *testing.T) {
		body := map[string]string{"email": "login@example.com"}
		w := doRequest(env, http.MethodPost, "/api/v1/auth/login", body, "")
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("401 – wrong password", func(t *testing.T) {
		body := map[string]string{
			"email":    "login@example.com",
			"password": "wrongpassword",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/auth/login", body, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("401 – non-existent user", func(t *testing.T) {
		body := map[string]string{
			"email":    "ghost@example.com",
			"password": "pass1234",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/auth/login", body, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestAuth_Refresh(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "refresh@example.com", "pass1234", "RefreshUser")

	t.Run("200 – valid token returns new tokens", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/auth/refresh", nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})

	t.Run("401 – no Authorization header", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/auth/refresh", nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("401 – malformed token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/auth/refresh", nil, "not.a.jwt")
		assertStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("401 – Bearer prefix missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/refresh", nil)
		req.Header.Set("Authorization", token) // missing "Bearer " prefix
		rec := httptest.NewRecorder()
		env.router.ServeHTTP(rec, req)
		assertStatus(t, rec, http.StatusUnauthorized)
	})
}
