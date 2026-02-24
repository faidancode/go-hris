package usererrors

import (
	"go-hris/internal/shared/apperror"
	"net/http"
)

var (
	ErrUserNotFound = apperror.New(
		apperror.CodeNotFound,
		"User not found",
		http.StatusNotFound,
	)

	ErrUserAlreadyExists = apperror.New(
		apperror.CodeConflict,
		"User with the same email already exists",
		http.StatusConflict,
	)

	ErrInvalidUserID = apperror.New(
		apperror.CodeInvalidInput,
		"Invalid user ID",
		http.StatusBadRequest,
	)

	ErrInvalidCompanyID = apperror.New(
		apperror.CodeInvalidInput,
		"Invalid company ID",
		http.StatusBadRequest,
	)

	ErrInvalidEmail = apperror.New(
		apperror.CodeInvalidInput,
		"Invalid email format",
		http.StatusBadRequest,
	)

	ErrMissingRequiredFields = apperror.New(
		apperror.CodeInvalidInput,
		"Missing required fields",
		http.StatusBadRequest,
	)

	ErrInvalidPassword = apperror.New(
		apperror.CodeInvalidInput,
		"Invalid password",
		http.StatusBadRequest,
	)

	ErrWrongPassword = apperror.New(
		apperror.CodeInvalidInput,
		"Current password is incorrect",
		http.StatusBadRequest,
	)

	ErrUserInactive = apperror.New(
		apperror.CodeForbidden,
		"User is inactive",
		http.StatusForbidden,
	)
)
