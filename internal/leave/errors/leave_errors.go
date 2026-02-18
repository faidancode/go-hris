package leaveerrors

import (
	"net/http"

	"go-hris/internal/shared/apperror"
)

var (
	ErrInvalidCompanyID = apperror.New(
		apperror.CodeInvalidInput,
		"invalid company id",
		http.StatusBadRequest,
	)
	ErrInvalidActorID = apperror.New(
		apperror.CodeInvalidInput,
		"invalid actor id",
		http.StatusBadRequest,
	)
	ErrInvalidEmployeeID = apperror.New(
		apperror.CodeInvalidInput,
		"invalid employee id",
		http.StatusBadRequest,
	)
	ErrInvalidDateFormat = apperror.New(
		apperror.CodeInvalidInput,
		"invalid date format, expected YYYY-MM-DD",
		http.StatusBadRequest,
	)
	ErrInvalidDateRange = apperror.New(
		apperror.CodeInvalidInput,
		"start_date must be before or equal end_date",
		http.StatusBadRequest,
	)
	ErrEmployeeNotInCompany = apperror.New(
		apperror.CodeInvalidInput,
		"employee does not belong to this company",
		http.StatusBadRequest,
	)
	ErrLeaveOverlap = apperror.New(
		apperror.CodeConflict,
		"leave already exists in overlapping period",
		http.StatusConflict,
	)
	ErrLeaveNotFound = apperror.New(
		apperror.CodeNotFound,
		"leave not found",
		http.StatusNotFound,
	)
	ErrInvalidStatusTransition = apperror.New(
		apperror.CodeInvalidState,
		"invalid leave status transition",
		http.StatusBadRequest,
	)
	ErrSubmittedDetailsImmutable = apperror.New(
		apperror.CodeInvalidState,
		"submitted leave details cannot be changed during approval",
		http.StatusBadRequest,
	)
	ErrApprovedByRequired = apperror.New(
		apperror.CodeInvalidInput,
		"approved_by is required when status is APPROVED",
		http.StatusBadRequest,
	)
	ErrInvalidApprovedBy = apperror.New(
		apperror.CodeInvalidInput,
		"invalid approved_by",
		http.StatusBadRequest,
	)
	ErrRejectionReasonRequired = apperror.New(
		apperror.CodeInvalidInput,
		"rejection_reason is required when status is REJECTED",
		http.StatusBadRequest,
	)
)
