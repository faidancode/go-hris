package companyerrors

import (
	"go-hris/internal/shared/apperror"
	"net/http"
)

var (
	ErrCompanyNotFound = apperror.New(
		apperror.CodeNotFound,
		"Company not found",
		http.StatusNotFound,
	)

	ErrCompanyAlreadyExists = apperror.New(
		apperror.CodeConflict,
		"Company with the same name already exists",
		http.StatusConflict,
	)

	ErrInvalidCompanyID = apperror.New(
		apperror.CodeInvalidInput,
		"Invalid company ID",
		http.StatusBadRequest,
	)

	ErrInvalidRegistrationType = apperror.New(
		apperror.CodeInvalidInput,
		"Invalid registration type",
		http.StatusBadRequest,
	)

	ErrRegistrationNotFound = apperror.New(
		apperror.CodeNotFound,
		"Company registration not found",
		http.StatusNotFound,
	)

	ErrRegistrationAlreadyExists = apperror.New(
		apperror.CodeConflict,
		"Registration type already exists for this company",
		http.StatusConflict,
	)

	ErrMissingRequiredFields = apperror.New(
		apperror.CodeInvalidInput,
		"Missing required fields",
		http.StatusBadRequest,
	)
)
