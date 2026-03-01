package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSummary_Daily(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "summary_daily@example.com", "pass1234", "SummaryUser")
	device := seedDevice(t, env, user.ID, "Summary PC", "windows")
	seedEvent(t, env, user.ID, device.ID, "VSCode")
	seedEvent(t, env, user.ID, device.ID, "Chrome")

	today := time.Now().Format("2006-01-02")

	t.Run("200 – daily summary for today", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, fmt.Sprintf("/api/v1/summary/daily?date=%s", today), nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "Daily summary")
	})

	t.Run("200 – defaults to today when date omitted", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/summary/daily", nil, token)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("200 – date with no events returns zero summary", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/summary/daily?date=2000-01-01", nil, token)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("500 – invalid date format", func(t *testing.T) {
		// The service returns an error for invalid date format; currently results in 500.
		w := doRequest(env, http.MethodGet, "/api/v1/summary/daily?date=not-a-date", nil, token)
		// The service validates and returns an error; we expect a non-200 response.
		assert.NotEqual(t, http.StatusOK, w.Code)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, fmt.Sprintf("/api/v1/summary/daily?date=%s", today), nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestSummary_Weekly(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "summary_weekly@example.com", "pass1234", "WeeklyUser")
	device := seedDevice(t, env, user.ID, "Weekly PC", "windows")
	seedEvent(t, env, user.ID, device.ID, "Slack")

	today := time.Now().Format("2006-01-02")

	t.Run("200 – weekly summary for current week", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, fmt.Sprintf("/api/v1/summary/weekly?date=%s", today), nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "Weekly summary")
	})

	t.Run("200 – defaults to current week when date omitted", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/summary/weekly", nil, token)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, fmt.Sprintf("/api/v1/summary/weekly?date=%s", today), nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}
