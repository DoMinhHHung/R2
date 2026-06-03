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

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func (e *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

var (
	ErrEmailExists        = New("EMAIL_EXISTS", "Email already registered", http.StatusConflict)
	ErrUserNotFound       = New("USER_NOT_FOUND", "User not found", http.StatusNotFound)
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrUserNotVerified    = New("USER_NOT_VERIFIED", "Please verify your email first", http.StatusForbidden)
	ErrInvalidOTP         = New("INVALID_OTP", "Invalid OTP code", http.StatusBadRequest)
	ErrOTPExpired         = New("OTP_EXPIRED", "OTP session expired, please start again", http.StatusGone)
	ErrMaxRetryExceeded   = New("MAX_RETRY_EXCEEDED", "Too many failed attempts", http.StatusTooManyRequests)
	ErrSessionNotFound    = New("SESSION_NOT_FOUND", "Session not found or expired", http.StatusUnauthorized)
	ErrTokenExpired       = New("TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized)
	ErrTokenInvalid       = New("TOKEN_INVALID", "Token is invalid", http.StatusUnauthorized)
	ErrResetTokenInvalid  = New("RESET_TOKEN_INVALID", "Reset token is invalid or expired", http.StatusBadRequest)
	ErrInternal           = New("INTERNAL_ERROR", "An internal error occurred", http.StatusInternalServerError)
	ErrForbidden          = New("FORBIDDEN", "You don't have permission to perform this action", http.StatusForbidden)
)
