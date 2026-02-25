package utils

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard response structure
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	logResponse(c, statusCode, message, nil)
	c.JSON(statusCode, response)
}

// OK sends a 200 OK response
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// Created sends a 201 Created response
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, errorCode string, message string, details string) {
	response := APIResponse{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    errorCode,
			Details: details,
		},
		Timestamp: time.Now().Unix(),
	}

	logResponse(c, statusCode, message, &details)
	c.JSON(statusCode, response)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, details string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message, "")
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message, "")
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message, "")
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string, details string) {
	Error(c, http.StatusConflict, "CONFLICT", message, details)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, details)
}

// NotImplemented sends a 501 Not Implemented response
func NotImplemented(c *gin.Context, message string) {
	Error(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", message, "")
}

// logResponse logs the response details
func logResponse(c *gin.Context, statusCode int, message string, errorDetails *string) {
	method := c.Request.Method
	path := c.Request.URL.Path
	clientIP := c.ClientIP()

	if errorDetails != nil && *errorDetails != "" {
		log.Printf("[%s] %s %s | Status: %d | IP: %s | Error: %s - %s",
			getStatusIcon(statusCode), method, path, statusCode, clientIP, message, *errorDetails)
	} else {
		log.Printf("[%s] %s %s | Status: %d | IP: %s | %s",
			getStatusIcon(statusCode), method, path, statusCode, clientIP, message)
	}
}

// getStatusIcon returns an icon based on status code range
func getStatusIcon(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "✓"
	case statusCode >= 400 && statusCode < 500:
		return "✗"
	case statusCode >= 500:
		return "!"
	default:
		return "•"
	}
}
