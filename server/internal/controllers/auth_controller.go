package controllers

import (
	"github.com/ayussh-2/timepad/internal/services"
	"github.com/ayussh-2/timepad/internal/utils"
	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=4"`
	Name     string `json:"name" binding:"required"`
}

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{
		service: service,
	}
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
