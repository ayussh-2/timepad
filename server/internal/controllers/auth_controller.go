package controllers

import (
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{
		service: service,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=4"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=4"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	params := services.RegisterUserParams{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	result, err := ac.service.RegisterUser(params)
	if err != nil {
		utils.Conflict(c, "Registration failed", err.Error())
		return
	}

	utils.Created(c, "User registered successfully", result)
}

func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	params := services.LoginUserParams{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := ac.service.LoginUser(params)

	if err != nil {
		print(err)
		utils.Unauthorized(c, "Login failed")
		return
	}

	utils.OK(c, "Login succefull", result)

}

func (ac *AuthController) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Validation failed", err.Error())
		return
	}

	result, err := ac.service.RefreshTokensFromToken(req.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	utils.OK(c, "Tokens refreshed successfully", result)
}

func (ac *AuthController) DeleteAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "User ID not found in context")
		return
	}

	if err := ac.service.DeleteAccount(userID.(string)); err != nil {
		utils.HandleError(c, "Failed to delete account", err)
		return
	}

	utils.OK(c, "Account deleted successfully", nil)
}
