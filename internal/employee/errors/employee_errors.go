package employeeerrors

import (
	"go-hris/internal/shared/apperror"
	"net/http"
)

var (
	ErrEmployeeNotFound = apperror.New(
		apperror.CodeNotFound,
		"Employee not found",
		http.StatusNotFound,
	)
	ErrEmployeeAlreadyExists = apperror.New(
		apperror.CodeConflict,
		"Employee with the same email already exists",
		http.StatusConflict,
	)
	ErrEmployeeNumberAlreadyExists = apperror.New(
		apperror.CodeConflict,
		"Employee number already exists in this company",
		http.StatusConflict,
	)
	ErrInvalidEmployeeID = apperror.New(
		apperror.CodeInvalidInput,
		"Invalid employee ID",
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
)
