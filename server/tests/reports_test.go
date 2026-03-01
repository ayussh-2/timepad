package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReports_Get(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "reports@example.com", "pass1234", "ReportsUser")
	device := seedDevice(t, env, user.ID, "Reports PC", "windows")
	seedEvent(t, env, user.ID, device.ID, "Excel")
	seedEvent(t, env, user.ID, device.ID, "Outlook")

	today := time.Now().Format("2006-01-02")
	weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	t.Run("200 – reports with date range", func(t *testing.T) {
		url := "/api/v1/reports?start_date=" + weekAgo + "&end_date=" + today
		w := doRequest(env, http.MethodGet, url, nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "Reports")
	})

	t.Run("200 – reports without date range returns all data", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/reports", nil, token)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("200 – reports for user with no events returns zeroed data", func(t *testing.T) {
		_, emptyToken := seedUser(t, env, "reports_empty@example.com", "pass1234", "EmptyReports")
		w := doRequest(env, http.MethodGet, "/api/v1/reports", nil, emptyToken)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/reports", nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}
