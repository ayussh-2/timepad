package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError is an error that carries an HTTP status code for proper response mapping.
type AppError struct {
	StatusCode int
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}

// NewNotFoundError creates a 404 AppError.
func NewNotFoundError(msg string) *AppError {
	return &AppError{StatusCode: http.StatusNotFound, Message: msg}
}

// NewBadRequestError creates a 400 AppError.
func NewBadRequestError(msg string) *AppError {
	return &AppError{StatusCode: http.StatusBadRequest, Message: msg}
}

// NewConflictError creates a 409 AppError.
func NewConflictError(msg string) *AppError {
	return &AppError{StatusCode: http.StatusConflict, Message: msg}
}

// HandleError inspects err for an *AppError and sends the appropriate HTTP response.
// If no AppError is found it falls back to a 500 Internal Server Error.
func HandleError(c *gin.Context, fallbackMsg string, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.StatusCode {
		case http.StatusNotFound:
			NotFound(c, appErr.Message)
		case http.StatusBadRequest:
			BadRequest(c, appErr.Message, "")
		case http.StatusConflict:
			Conflict(c, appErr.Message, "")
		default:
			InternalServerError(c, fallbackMsg, err)
		}
		return
	}
	InternalServerError(c, fallbackMsg, err)
}
