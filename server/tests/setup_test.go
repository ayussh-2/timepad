// Package tests contains integration tests that exercise the full HTTP stack
// (router → middleware → controller → service → SQLite in-memory DB).
package tests

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ayussh-2/timepad/internal/controllers"
	"github.com/ayussh-2/timepad/internal/middleware"
	"github.com/ayussh-2/timepad/internal/models"
	"github.com/ayussh-2/timepad/internal/routes"
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testEnv holds shared test state reused across subtests.
type testEnv struct {
	db      *gorm.DB
	jwtUtil *utils.JWTUtil
	router  *gin.Engine
}

// setupTestEnv creates an in-memory SQLite database, migrates all models,
// generates ephemeral RSA keys, and wires up the full Gin router.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Generate in-memory RSA keys so tests need no key files on disk.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "generate RSA key")

	jwtUtil := utils.NewJWTUtilFromKeys(privateKey, &privateKey.PublicKey)

	// Open SQLite in-memory database.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "open sqlite db")

	// Migrate all models.
	err = db.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.Category{},
		&models.ActivityEvent{},
		&models.UserSetting{},
	)
	require.NoError(t, err, "auto migrate")

	gin.SetMode(gin.TestMode)
	router := buildTestRouter(db, jwtUtil)

	return &testEnv{db: db, jwtUtil: jwtUtil, router: router}
}

// buildTestRouter constructs the Gin router exactly like routes.SetupRouter
// but uses the provided test db and jwtUtil.
func buildTestRouter(db *gorm.DB, jwtUtil *utils.JWTUtil) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	healthSvc := services.NewHealthService()
	authSvc := services.NewAuthService(db, jwtUtil)
	eventsSvc := services.NewEventsService(db)
	summarySvc := services.NewSummaryService(db)
	reportsSvc := services.NewReportsService(db)
	categoriesSvc := services.NewCategoriesService(db)
	devicesSvc := services.NewDevicesService(db)
	settingsSvc := services.NewSettingsService(db)

	healthCtrl := controllers.NewHealthController(healthSvc)
	authCtrl := controllers.NewAuthController(authSvc)
	eventsCtrl := controllers.NewEventsController(eventsSvc)
	summaryCtrl := controllers.NewSummaryController(summarySvc)
	reportsCtrl := controllers.NewReportsController(reportsSvc)
	categoriesCtrl := controllers.NewCategoriesController(categoriesSvc)
	devicesCtrl := controllers.NewDevicesController(devicesSvc)
	settingsCtrl := controllers.NewSettingsController(settingsSvc)

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	routes.RegisterAuthRoutes(r, auth, authCtrl, jwtUtil)
	routes.RegisterHealthRoutes(r, v1, healthCtrl)

	protected := v1.Group("/")
	protected.Use(middleware.Auth(jwtUtil))
	{
		routes.RegisterEventsRoutes(protected, eventsCtrl)
		routes.RegisterSummaryRoutes(protected, summaryCtrl)
		routes.RegisterReportsRoutes(protected, reportsCtrl)
		routes.RegisterCategoriesRoutes(protected, categoriesCtrl)
		routes.RegisterDevicesRoutes(protected, devicesCtrl)
		routes.RegisterSettingsRoutes(protected, settingsCtrl)
	}

	return r
}

// ─── Seed helpers ─────────────────────────────────────────────────────────────

// seedUser inserts a user with the given email/password and creates default settings.
// It returns the user and a valid access token.
func seedUser(t *testing.T, env *testEnv, email, password, name string) (models.User, string) {
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

	// Create default settings for the seeded user.
	settings := models.UserSetting{
		UserID:            user.ID,
		IdleThreshold:     300,
		TrackingEnabled:   true,
		DataRetentionDays: 365,
		ExcludedApps:      pq.StringArray{},
		ExcludedUrls:      pq.StringArray{},
	}
	require.NoError(t, env.db.Create(&settings).Error)

	token, err := env.jwtUtil.GenerateAccessToken(user.ID.String())
	require.NoError(t, err)

	return user, token
}

// seedDevice inserts a device for the given user and returns it.
func seedDevice(t *testing.T, env *testEnv, userID uuid.UUID, name, platform string) models.Device {
	t.Helper()

	device := models.Device{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Platform:  platform,
		DeviceKey: fmt.Sprintf("%s-%s", platform, uuid.New().String()),
	}
	require.NoError(t, env.db.Create(&device).Error)
	return device
}

// seedCategory inserts a category for the given user and returns it.
func seedCategory(t *testing.T, env *testEnv, userID uuid.UUID, name string) models.Category {
	t.Helper()

	cat := models.Category{
		ID:     uuid.New(),
		UserID: &userID,
		Name:   name,
		Color:  "#FF0000",
	}
	require.NoError(t, env.db.Create(&cat).Error)
	return cat
}

// seedEvent inserts an activity event for the given user/device and returns it.
func seedEvent(t *testing.T, env *testEnv, userID, deviceID uuid.UUID, appName string) models.ActivityEvent {
	t.Helper()

	now := time.Now().UTC()
	event := models.ActivityEvent{
		ID:           uuid.New(),
		UserID:       userID,
		DeviceID:     deviceID,
		AppName:      appName,
		WindowTitle:  "Test Window",
		StartTime:    now.Add(-5 * time.Minute),
		EndTime:      now,
		DurationSecs: 300,
		IsIdle:       false,
	}
	require.NoError(t, env.db.Create(&event).Error)
	return event
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────────

// doRequest performs an HTTP request against the test router and returns the recorder.
func doRequest(env *testEnv, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	return w
}

// parseBody unmarshals the response body into v.
func parseBody(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	require.NoError(t, json.NewDecoder(w.Body).Decode(v))
}

// assertStatus asserts the HTTP status code equals expected.
func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("expected status %d, got %d\nbody: %s", expected, w.Code, w.Body.String())
	}
}

// apiResp is the generic response envelope returned by all endpoints.
type apiResp struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Details string `json:"details"`
	} `json:"error"`
}
