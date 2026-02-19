package employeesalary

import (
	"errors"
	"strings"

	employeesalaryerrors "go-hris/internal/employeesalary/errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" && pgErr.ConstraintName == "uq_employee_salary_effective" {
			return employeesalaryerrors.ErrSalaryEffectiveDateAlreadyExists
		}
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "duplicate key value") && strings.Contains(errMsg, "uq_employee_salary_effective") {
		return employeesalaryerrors.ErrSalaryEffectiveDateAlreadyExists
	}

	return err
}
