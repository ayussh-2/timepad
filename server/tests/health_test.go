package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth_GetHealth(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("200 – health endpoint returns ok", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/health", nil, "")
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "healthy")
	})
}

func TestHealth_Ping(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("200 – ping endpoint returns pong", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/ping", nil, "")
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})
}
