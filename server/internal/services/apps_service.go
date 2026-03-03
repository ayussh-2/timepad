package services

import (
	"errors"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppsService struct {
	db *gorm.DB
}

func NewAppsService(db *gorm.DB) *AppsService {
	return &AppsService{db: db}
}

// ListApps returns all tracked apps for a user, each with its assigned category.
func (s *AppsService) ListApps(userID string) ([]models.App, error) {
	var apps []models.App
	err := s.db.Where("user_id = ?", userID).
		Preload("Category").
		Order("last_seen_at desc").
		Find(&apps).Error
	if err != nil {
		return nil, errors.New("failed to fetch apps")
	}
	return apps, nil
}

// SetAppCategory directly assigns (or clears) the category for an app.
func (s *AppsService) SetAppCategory(userID, appID string, categoryID *string) (*models.App, error) {
	var app models.App
	if err := s.db.Where("id = ? AND user_id = ?", appID, userID).First(&app).Error; err != nil {
		return nil, utils.NewNotFoundError("app not found")
	}

	if categoryID == nil || *categoryID == "" {
		if err := s.db.Model(&app).Update("category_id", nil).Error; err != nil {
			return nil, errors.New("failed to clear app category")
		}
	} else {
		catID, err := uuid.Parse(*categoryID)
		if err != nil {
			return nil, errors.New("invalid category ID")
		}
		var cat models.Category
		if err := s.db.Where("id = ? AND (user_id = ? OR is_system = true)", catID, userID).First(&cat).Error; err != nil {
			return nil, utils.NewNotFoundError("category not found")
		}
		if err := s.db.Model(&app).Update("category_id", catID).Error; err != nil {
			return nil, errors.New("failed to set app category")
		}
	}

	// Reload with category preloaded.
	s.db.Where("id = ?", app.ID).Preload("Category").First(&app)
	return &app, nil
}

// SetAppSystem marks or unmarks an app as a system app.
// System apps are shown with a visual indicator and excluded from productivity stats.
func (s *AppsService) SetAppSystem(userID, appID string, isSystem bool) (*models.App, error) {
	var app models.App
	if err := s.db.Where("id = ? AND user_id = ?", appID, userID).First(&app).Error; err != nil {
		return nil, utils.NewNotFoundError("app not found")
	}
	if err := s.db.Model(&app).Update("is_system", isSystem).Error; err != nil {
		return nil, errors.New("failed to update app system flag")
	}
	s.db.Where("id = ?", app.ID).Preload("Category").First(&app)
	return &app, nil
}

// ClassifyApp finds-or-creates a default category matching is_productive and
// assigns it to the app. Pass nil isProductive to clear the category (neutral).
func (s *AppsService) ClassifyApp(userID, appID string, isProductive *bool) (*models.App, error) {
	parsedUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	var app models.App
	if err := s.db.Where("id = ? AND user_id = ?", appID, userID).First(&app).Error; err != nil {
		return nil, utils.NewNotFoundError("app not found")
	}

	if isProductive == nil {
		if err := s.db.Model(&app).Update("category_id", nil).Error; err != nil {
			return nil, errors.New("failed to clear app classification")
		}
		app.CategoryID = nil
		app.Category = nil
		return &app, nil
	}

	// Find an existing user category with the matching productivity level.
	var cat models.Category
	err = s.db.Where("user_id = ? AND is_productive = ?", parsedUID, *isProductive).First(&cat).Error
	if err != nil {
		// Create a simple default category.
		name := "Productive"
		color := "#7a9a6d"
		if !*isProductive {
			name = "Distraction"
			color = "#c4a77d"
		}
		cat = models.Category{
			UserID:       &parsedUID,
			Name:         name,
			Color:        color,
			IsProductive: isProductive,
		}
		if err := s.db.Create(&cat).Error; err != nil {
			return nil, errors.New("failed to create default category")
		}
	}

	if err := s.db.Model(&app).Update("category_id", cat.ID).Error; err != nil {
		return nil, errors.New("failed to classify app")
	}
	app.CategoryID = &cat.ID
	app.Category = &cat
	return &app, nil
}
