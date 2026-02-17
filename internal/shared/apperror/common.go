package apperror

import "net/http"

var (
	ErrNotFound = New(
		CodeNotFound,
		"Resource not found",
		http.StatusNotFound,
	)

	ErrForbidden = New(
		CodeForbidden,
		"You do not have permission to access this resource",
		http.StatusForbidden,
	)

	ErrInternal = New(
		CodeInternalError,
		"An unexpected error occurred",
		http.StatusInternalServerError,
	)

	ErrUnauthorized = New(
		CodeUnauthorized,
		"Authentication is required",
		http.StatusUnauthorized,
	)

	ErrInvalidInput = New(
		CodeInvalidInput,
		"The provided input is invalid",
		http.StatusBadRequest,
	)
)
