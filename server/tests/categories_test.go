package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategories_Get(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "cat_get@example.com", "pass1234", "CatGetUser")
	seedCategory(t, env, user.ID, "Work")
	seedCategory(t, env, user.ID, "Personal")

	t.Run("200 – returns user categories", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/categories", nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "fetched")
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		w := doRequest(env, http.MethodGet, "/api/v1/categories", nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestCategories_Create(t *testing.T) {
	env := setupTestEnv(t)
	_, token := seedUser(t, env, "cat_create@example.com", "pass1234", "CatCreateUser")

	t.Run("201 – create valid category", func(t *testing.T) {
		body := map[string]interface{}{
			"name":          "Entertainment",
			"color":         "#FF5733",
			"is_productive": false,
		}
		w := doRequest(env, http.MethodPost, "/api/v1/categories", body, token)
		assertStatus(t, w, http.StatusCreated)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
		assert.Contains(t, resp.Message, "created")
	})

	t.Run("201 – create category with default color", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Study",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/categories", body, token)
		assertStatus(t, w, http.StatusCreated)
	})

	t.Run("400 – missing name field", func(t *testing.T) {
		body := map[string]interface{}{
			"color": "#AABBCC",
		}
		w := doRequest(env, http.MethodPost, "/api/v1/categories", body, token)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		body := map[string]interface{}{"name": "Social"}
		w := doRequest(env, http.MethodPost, "/api/v1/categories", body, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestCategories_Update(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "cat_update@example.com", "pass1234", "CatUpdateUser")
	cat := seedCategory(t, env, user.ID, "UpdateMe")

	t.Run("200 – update category name and color", func(t *testing.T) {
		newName := "Updated Name"
		body := map[string]interface{}{
			"name":  newName,
			"color": "#123456",
		}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/categories/%s", cat.ID), body, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})

	t.Run("200 – update is_productive flag", func(t *testing.T) {
		prod := true
		body := map[string]interface{}{"is_productive": prod}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/categories/%s", cat.ID), body, token)
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("404 – category belongs to another user", func(t *testing.T) {
		_, otherToken := seedUser(t, env, "catother@example.com", "pass1234", "Other")
		body := map[string]interface{}{"name": "Hacked"}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/categories/%s", cat.ID), body, otherToken)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("404 – non-existent category ID", func(t *testing.T) {
		fakeID := "00000000-0000-0000-0000-000000000000"
		body := map[string]interface{}{"name": "Ghost"}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/categories/%s", fakeID), body, token)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		body := map[string]interface{}{"name": "NoAuth"}
		w := doRequest(env, http.MethodPatch, fmt.Sprintf("/api/v1/categories/%s", cat.ID), body, "")
		assertStatus(t, w, http.StatusUnauthorized)
	})
}

func TestCategories_Delete(t *testing.T) {
	env := setupTestEnv(t)
	user, token := seedUser(t, env, "cat_delete@example.com", "pass1234", "CatDeleteUser")

	t.Run("200 – delete own category", func(t *testing.T) {
		cat := seedCategory(t, env, user.ID, "DeleteMe")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%s", cat.ID), nil, token)
		assertStatus(t, w, http.StatusOK)

		var resp apiResp
		parseBody(t, w, &resp)
		assert.True(t, resp.Success)
	})

	t.Run("404 – category already deleted", func(t *testing.T) {
		cat := seedCategory(t, env, user.ID, "DeleteMeTwice")
		doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%s", cat.ID), nil, token)
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%s", cat.ID), nil, token)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("404 – category belongs to another user", func(t *testing.T) {
		cat := seedCategory(t, env, user.ID, "Protected")
		_, otherToken := seedUser(t, env, "catdelother@example.com", "pass1234", "Other")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%s", cat.ID), nil, otherToken)
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("401 – no auth token", func(t *testing.T) {
		cat := seedCategory(t, env, user.ID, "NoAuthDelete")
		w := doRequest(env, http.MethodDelete, fmt.Sprintf("/api/v1/categories/%s", cat.ID), nil, "")
		assertStatus(t, w, http.StatusUnauthorized)
		require.NotNil(t, cat.ID)
	})
}
