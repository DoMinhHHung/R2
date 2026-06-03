package apperr

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

var (
	ErrUserNotFound      = New("USER_NOT_FOUND", "User not found", http.StatusNotFound)
	ErrUserBanned        = New("USER_BANNED", "User is banned", http.StatusForbidden)
	ErrUserDeleted       = New("USER_DELETED", "User account no longer exists", http.StatusGone)
	ErrUserAlreadyExists = New("USER_ALREADY_EXISTS", "User already exists", http.StatusConflict)
	ErrPhoneExists       = New("PHONE_EXISTS", "Phone number already in use", http.StatusConflict)
	ErrForbidden         = New("FORBIDDEN", "You don't have permission", http.StatusForbidden)
	ErrUnauthorized      = New("UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
	ErrInvalidInput      = New("INVALID_INPUT", "Invalid input data", http.StatusBadRequest)
	ErrFileTooLarge      = New("FILE_TOO_LARGE", "File size exceeds limit", http.StatusRequestEntityTooLarge)
	ErrInvalidFileType   = New("INVALID_FILE_TYPE", "File type not allowed", http.StatusBadRequest)
	ErrNoAvatar          = New("NO_AVATAR", "User has no avatar to delete", http.StatusBadRequest)
	ErrInternal          = New("INTERNAL_ERROR", "An internal error occurred", http.StatusInternalServerError)
)
