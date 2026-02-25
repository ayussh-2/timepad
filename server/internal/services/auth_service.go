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
	RefreshToken string
	AccessToken  string
}

type RegisterUserParams struct {
	Name     string
	Email    string
	Password string
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
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
