package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvents_Ingest(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "events_ingest@example.com", "pass1234", "IngestUser")
	device := seedDevice(t, env, user.ID, "My PC", "windows")

	validEvent := map[string]interface{}{
		"app_name":     "VSCode",
		"window_title": "main.go",
		"start_time":   time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		"end_time":     time.Now().Format(time.RFC3339),
		"is_idle":      false,
	}

	t.Run("201 – ingest valid events", func(t *testing.T) {
		body := map[string]interface{}{
			"device_key": device.DeviceKey,
			"events":     []interface{}{validEvent},
		}
		w := doRequest(env, http.MethodPost, "/api/v1/events", body, token)
		assertStatus(t, w, http.StatusCreated)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "ingested")
	})

	t.Run("400 – missing device_key", func(t *testing.T) {
		body := map[string]interface{}{
			"events": []interface{}{validEvent},
		}
		w := doRequest(env, http.MethodPost, "/api/v1/events", body, token)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("400 – empty events array", func(t *testing.T) {
		body := map[string]interface{}{
			"device_key": device.DeviceKey,
			"events":     []interface{}{},
		}
		w := doRequest(env, http.MethodPost, "/api/v1/events", body, token)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("404 – unknown device key", func(t *testing.T) {
		body := map[string]interface{}{
			"device_key": "windows-00000000-0000-0000-0000-000000000000",
			"events":     []interface{}{validEvent},
		}
		w := doRequest(env, http.MethodPost, "/api/v1/events", body, token)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		body := map[string]interface{}{
			"device_key": device.DeviceKey,
			"events":     []interface{}{validEvent},
		}
		w := doRequest(env, http.MethodPost, "/api/v1/events", body, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestEvents_Get(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "events_get@example.com", "pass1234", "GetUser")
	device := seedDevice(t, env, user.ID, "Laptop", "windows")
	seedEvent(t, env, user.ID, device.ID, "Chrome")
	seedEvent(t, env, user.ID, device.ID, "Slack")

	t.Run("200 – returns events list", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/events", nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})

	t.Run("200 – respects limit query param", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/events?limit=1&offset=0", nil, token)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/events", nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestEvents_Timeline(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "events_timeline@example.com", "pass1234", "TimelineUser")
	device := seedDevice(t, env, user.ID, "PC", "windows")
	seedEvent(t, env, user.ID, device.ID, "Firefox")

	today := time.Now().Format("2006-01-02")

	t.Run("200 – returns timeline for date", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, fmt.Sprintf("/api/v1/timeline?date=%s", today), nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})

	t.Run("400 – missing date param", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/timeline", nil, token)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, fmt.Sprintf("/api/v1/timeline?date=%s", today), nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestEvents_Edit(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "events_edit@example.com", "pass1234", "EditUser")
	device := seedDevice(t, env, user.ID, "Edit PC", "windows")
	event := seedEvent(t, env, user.ID, device.ID, "Notepad")

	t.Run("200 – mark event as private", func(t *testing.T) {
		priv := true
		body := map[string]interface{}{"is_private": priv}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/events/%s", event.ID), body, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})

	t.Run("200 – clear category", func(t *testing.T) {
		catID := ""
		body := map[string]interface{}{"category_id": catID}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/events/%s", event.ID), body, token)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("404 – non-existent event ID", func(t *testing.T) {
		body := map[string]interface{}{"is_private": true}
		fakeID := "00000000-0000-0000-0000-000000000000"
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/events/%s", fakeID), body, token)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("404 – event belongs to another user", func(t *testing.T) {
		_, otherToken := seedUser(t, env, "otheredit@example.com", "pass1234", "Other")
		body := map[string]interface{}{"is_private": true}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/events/%s", event.ID), body, otherToken)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		body := map[string]interface{}{"is_private": true}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/events/%s", event.ID), body, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestEvents_Delete(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "events_delete@example.com", "pass1234", "DeleteUser")
	device := seedDevice(t, env, user.ID, "Del PC", "windows")

	t.Run("200 – delete own event", func(t *testing.T) {
		event := seedEvent(t, env, user.ID, device.ID, "Word")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/events/%s", event.ID), nil, token)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("404 – already deleted event", func(t *testing.T) {
		event := seedEvent(t, env, user.ID, device.ID, "Excel")
		// First delete succeeds.
		doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/events/%s", event.ID), nil, token)
		// Second delete should be 404.
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/events/%s", event.ID), nil, token)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("404 – event belongs to another user", func(t *testing.T) {
		event := seedEvent(t, env, user.ID, device.ID, "PowerPoint")
		_, otherToken := seedUser(t, env, "otherdel@example.com", "pass1234", "Other")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/events/%s", event.ID), nil, otherToken)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		event := seedEvent(t, env, user.ID, device.ID, "Outlook")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/events/%s", event.ID), nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
		require.NotNil(t, event.ID)
	})
}
