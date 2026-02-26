package services

import (
	"errors"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoriesService struct {
	db *gorm.DB
}

func NewCategoriesService(db *gorm.DB) *CategoriesService {
	return &CategoriesService{
		db: db,
	}
}

func (s *CategoriesService) GetCategories(userID string) ([]models.Category, error) {
	var categories []models.Category

	// Fetch user-specific categories OR system-wide categories
	err := s.db.Where("user_id = ? OR is_system = ?", userID, true).Find(&categories).Error
	if err != nil {
		return nil, errors.New("failed to fetch categories")
	}

	return categories, nil
}

type CreateCategoryParams struct {
	Name         string `json:"name" binding:"required"`
	Color        string `json:"color"`
	Icon         string `json:"icon"`
	IsProductive *bool  `json:"is_productive"`
}

func (s *CategoriesService) CreateCategory(userID string, params CreateCategoryParams) (*models.Category, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	color := params.Color
	if color == "" {
		color = "#6B7280"
	}

	category := models.Category{
		UserID:       &parsedUserID,
		Name:         params.Name,
		Color:        color,
		Icon:         params.Icon,
		IsSystem:     false,
		IsProductive: params.IsProductive,
	}

	if err := s.db.Create(&category).Error; err != nil {
		return nil, errors.New("failed to create category")
	}

	return &category, nil
}

type UpdateCategoryParams struct {
	Name         *string `json:"name"`
	Color        *string `json:"color"`
	Icon         *string `json:"icon"`
	IsProductive *bool   `json:"is_productive"`
}

func (s *CategoriesService) UpdateCategory(userID string, categoryID string, params UpdateCategoryParams) error {
	var category models.Category
	if err := s.db.Where("id = ? AND user_id = ?", categoryID, userID).First(&category).Error; err != nil {
		return errors.New("category not found or unauthorized")
	}

	updates := map[string]interface{}{}
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.Color != nil {
		updates["color"] = *params.Color
	}
	if params.Icon != nil {
		updates["icon"] = *params.Icon
	}
	if params.IsProductive != nil {
		updates["is_productive"] = *params.IsProductive
	}

	if len(updates) == 0 {
		return nil
	}

	if err := s.db.Model(&category).Updates(updates).Error; err != nil {
		return errors.New("failed to update category")
	}

	return nil
}

func (s *CategoriesService) DeleteCategory(userID string, categoryID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Nullify category_id on events that reference this category
		if err := tx.Model(&models.ActivityEvent{}).
			Where("category_id = ? AND user_id = ?", categoryID, userID).
			Update("category_id", nil).Error; err != nil {
			return errors.New("failed to unlink events from category")
		}

		result := tx.Where("id = ? AND user_id = ?", categoryID, userID).Delete(&models.Category{})
		if result.Error != nil {
			return errors.New("failed to delete category")
		}
		if result.RowsAffected == 0 {
			return errors.New("category not found or unauthorized")
		}
		return nil
	})
}
