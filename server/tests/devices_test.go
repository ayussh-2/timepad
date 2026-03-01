package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevices_Get(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "dev_get@example.com", "pass1234", "DevGetUser")
	seedDevice(t, env, user.ID, "Work Laptop", "windows")
	seedDevice(t, env, user.ID, "Phone", "android")

	t.Run("200 – returns user devices", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/devices", nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "fetched")
	})

	t.Run("200 – returns empty list for user with no devices", func(t *testing.T) {
		_, noDevToken := seedUser(t, env, "dev_empty@example.com", "pass1234", "EmptyUser")
		w := doRequest(env, http.MethodGet, "/api/v1/devices", nil, noDevToken)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/devices", nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestDevices_Register(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "dev_register@example.com", "pass1234", "DevRegUser")

	t.Run("201 – register windows device", func(t *testing.T) {
		body := map[string]string{
			"name":     "My PC",
			"platform": "windows",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/devices", body, token)
		assertStatus(t, w, http.StatusCreated)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "registered")
	})

	t.Run("201 – register android device", func(t *testing.T) {
		body := map[string]string{
			"name":     "Pixel 8",
			"platform": "android",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/devices", body, token)
		assertStatus(t, w, http.StatusCreated)
	})

	t.Run("201 – register browser device", func(t *testing.T) {
		body := map[string]string{
			"name":     "Chrome Extension",
			"platform": "browser",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/devices", body, token)
		assertStatus(t, w, http.StatusCreated)
	})

	t.Run("400 – missing name", func(t *testing.T) {
		body := map[string]string{"platform": "windows"}
		w := doRequest(env, http.MethodPost, "/api/v1/devices", body, token)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("400 – invalid platform value", func(t *testing.T) {
		body := map[string]string{
			"name":     "iPad",
			"platform": "ios", // not allowed
		}
		w := doRequest(env, http.MethodPost, "/api/v1/devices", body, token)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		body := map[string]string{"name": "Ghost", "platform": "windows"}
		w := doRequest(env, http.MethodPost, "/api/v1/devices", body, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestDevices_Delete(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "dev_delete@example.com", "pass1234", "DevDelUser")

	t.Run("200 – delete own device", func(t *testing.T) {
		device := seedDevice(t, env, user.ID, "Delete Me", "windows")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/devices/%s", device.ID), nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})

	t.Run("404 – device already deleted", func(t *testing.T) {
		device := seedDevice(t, env, user.ID, "Delete Twice", "windows")
		doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/devices/%s", device.ID), nil, token)
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/devices/%s", device.ID), nil, token)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("404 – device belongs to another user", func(t *testing.T) {
		device := seedDevice(t, env, user.ID, "Protected Device", "windows")
		_, otherToken := seedUser(t, env, "devother@example.com", "pass1234", "Other")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/devices/%s", device.ID), nil, otherToken)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		device := seedDevice(t, env, user.ID, "NoAuth Device", "windows")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/devices/%s", device.ID), nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
		require.NotNil(t, device.ID)
	})
}
