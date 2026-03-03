package services

import (
	"errors"

	"github.com/ayussh-2/timepad/internal/models"
	"github.com/ayussh-2/timepad/internal/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	db      *gorm.DB
	jwtUtil *utils.JWTUtil
}

func NewAuthService(db *gorm.DB, jwtUtil *utils.JWTUtil) *AuthService {
	return &AuthService{
		db:      db,
		jwtUtil: jwtUtil,
	}
}

type RegisterUserResult struct {
	UserId       string
	Name         string
	Email        string
	RefreshToken string
	AccessToken  string
}

type RegisterUserParams struct {
	Name     string
	Email    string
	Password string
}

type LoginUserParams struct {
	Email    string
	Password string
}

type LoginUserResult = RegisterUserResult

type RefreshTokensResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) RegisterUser(params RegisterUserParams) (*RegisterUserResult, error) {

	var existingUser models.User
	if result := s.db.Where("email = ?", params.Email).First(&existingUser); result.Error == nil {
		return nil, errors.New("Email already in use")
	}

	pwdHash, err := utils.HashPassword(params.Password)
	if err != nil {
		return nil, errors.New("Failed to hash password!")
	}

	user := models.User{
		DisplayName:  params.Name,
		Email:        params.Email,
		PasswordHash: pwdHash,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, errors.New("Fail to create new user!")
	}

	// Create default settings for the new user.
	defaultSettings := models.UserSetting{
		UserID:            user.ID,
		IdleThreshold:     300,
		TrackingEnabled:   true,
		DataRetentionDays: 365,
	}
	// Ignore error – settings can be created later via PUT /settings.
	_ = s.db.Create(&defaultSettings).Error

	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &RegisterUserResult{
		UserId:       user.ID.String(),
		Name:         user.DisplayName,
		Email:        user.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) LoginUser(params LoginUserParams) (*LoginUserResult, error) {
	var user models.User

	if err := s.db.Where("email = ?", params.Email).First(&user).Error; err != nil {
		return nil, errors.New("User not found!")
	}

	isValidPassword := utils.CheckPassword(params.Password, user.PasswordHash)

	if !isValidPassword {
		return nil, errors.New("Invalid credentials!")
	}

	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &LoginUserResult{
		UserId:       user.ID.String(),
		Name:         user.DisplayName,
		Email:        user.Email,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshTokens(userID string) (*RefreshTokensResult, error) {
	newAccessToken, err := s.jwtUtil.GenerateAccessToken(userID)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	newRefreshToken, err := s.jwtUtil.GenerateRefreshToken(userID)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &RefreshTokensResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// DeleteAccount permanently removes the user and all their associated data
// via the CASCADE constraints defined on the database relations.
func (s *AuthService) DeleteAccount(userID string) error {
	result := s.db.Where("id = ?", userID).Delete(&models.User{})
	if result.Error != nil {
		return errors.New("failed to delete account")
	}
	if result.RowsAffected == 0 {
		return utils.NewNotFoundError("user not found")
	}
	return nil
}
