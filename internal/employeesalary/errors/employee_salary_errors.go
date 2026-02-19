package employeesalaryerrors

import (
	"go-hris/internal/shared/apperror"
	"net/http"
)

var (
	ErrSalaryEffectiveDateAlreadyExists = apperror.New(
		apperror.CodeConflict,
		"Salary for this employee and effective date already exists",
		http.StatusConflict,
	)
)
