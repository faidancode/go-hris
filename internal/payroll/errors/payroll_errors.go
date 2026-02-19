package payrollerrors

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
	ErrInvalidDepartmentID = apperror.New(
		apperror.CodeInvalidInput,
		"invalid department id",
		http.StatusBadRequest,
	)
	ErrInvalidDateFormat = apperror.New(
		apperror.CodeInvalidInput,
		"invalid date format, expected YYYY-MM-DD",
		http.StatusBadRequest,
	)
	ErrInvalidPeriodFormat = apperror.New(
		apperror.CodeInvalidInput,
		"invalid period format, expected YYYY-MM",
		http.StatusBadRequest,
	)
	ErrInvalidDateRange = apperror.New(
		apperror.CodeInvalidInput,
		"period_start must be before or equal period_end",
		http.StatusBadRequest,
	)
	ErrEmployeeNotInCompany = apperror.New(
		apperror.CodeInvalidInput,
		"employee does not belong to this company",
		http.StatusBadRequest,
	)
	ErrPayrollOverlap = apperror.New(
		apperror.CodeConflict,
		"payroll already exists in overlapping period",
		http.StatusConflict,
	)
	ErrPayrollNotFound = apperror.New(
		apperror.CodeNotFound,
		"payroll not found",
		http.StatusNotFound,
	)
	ErrInvalidStatusTransition = apperror.New(
		apperror.CodeInvalidState,
		"invalid payroll status transition",
		http.StatusBadRequest,
	)
	ErrRegenerateOnlyDraft = apperror.New(
		apperror.CodeInvalidState,
		"payroll can only be regenerated while status is DRAFT",
		http.StatusBadRequest,
	)
	ErrDeleteOnlyDraft = apperror.New(
		apperror.CodeInvalidState,
		"payroll can only be deleted while status is DRAFT",
		http.StatusBadRequest,
	)
	ErrInvalidMoneyValue = apperror.New(
		apperror.CodeInvalidInput,
		"salary component values cannot be negative",
		http.StatusBadRequest,
	)
	ErrInvalidComponent = apperror.New(
		apperror.CodeInvalidInput,
		"invalid payroll component payload",
		http.StatusBadRequest,
	)
	ErrInvalidStatusFilter = apperror.New(
		apperror.CodeInvalidInput,
		"invalid payroll status filter",
		http.StatusBadRequest,
	)
	ErrPayslipNotGenerated = apperror.New(
		apperror.CodeNotFound,
		"payslip is not generated yet",
		http.StatusNotFound,
	)
)
