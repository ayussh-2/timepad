package tests

import (
	"net/http"
	"testing"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedUserNoSettings creates a user WITHOUT default settings for testing the 404 path.
func seedUserNoSettings(t *testing.T, env *testEnv, email, password, name string) (models.User, string) {
	t.Helper()

	hash, err := utils.HashPassword(password)
	require.NoError(t, err)

	user := models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		DisplayName:  name,
		Timezone:     "UTC",
	}
	require.NoError(t, env.db.Create(&user).Error)

	token, err := env.jwtUtil.GenerateAccessToken(user.ID.String())
	require.NoError(t, err)

	return user, token
}

func TestSettings_Get(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "settings@example.com", "pass1234", "SettingsUser")

	t.Run("200 – settings exist", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/settings", nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "fetched")
	})

	t.Run("200 – default settings returned for user without defaults", func(t *testing.T) {
		_, noSettingsToken := seedUserNoSettings(t, env, "nosettings@example.com", "pass1234", "NoSettings")

		w := doRequest(env, http.MethodGet, "/api/v1/settings", nil, noSettingsToken)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "fetched")
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/settings", nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestSettings_Update(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "settingsupdate@example.com", "pass1234", "UpdateUser")

	t.Run("200 – update existing settings", func(t *testing.T) {
		idle := 600
		body := map[string]interface{}{
			"idle_threshold":   idle,
			"tracking_enabled": false,
		}
		w := doRequest(env, http.MethodPut, "/api/v1/settings", body, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})

	t.Run("200 – upsert creates settings when absent", func(t *testing.T) {
		_, noSettingsToken := seedUserNoSettings(t, env, "upsert@example.com", "pass1234", "UpsertUser")

		body := map[string]interface{}{
			"data_retention_days": 30,
		}
		w := doRequest(env, http.MethodPut, "/api/v1/settings", body, noSettingsToken)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("200 – update excluded apps and urls", func(t *testing.T) {
		body := map[string]interface{}{
			"excluded_apps": []string{"Slack", "Discord"},
			"excluded_urls": []string{"reddit.com"},
		}
		w := doRequest(env, http.MethodPut, "/api/v1/settings", body, token)
		assertStatus(t, w, http.StatusOK)

		// Verify the change is readable via GET.
		wGet := doRequest(env, http.MethodGet, "/api/v1/settings", nil, token)
		assertStatus(t, wGet, http.StatusOK)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodPut, "/api/v1/settings", map[string]interface{}{}, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}
