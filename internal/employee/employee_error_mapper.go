package employee

import (
	"errors"
	"strings"

	employeeerrors "go-hris/internal/employee/errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func mapRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return employeeerrors.ErrEmployeeNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "uq_employee_number":
				return employeeerrors.ErrEmployeeNumberAlreadyExists
			case "uq_employee_email":
				return employeeerrors.ErrEmployeeAlreadyExists
			}
		}
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "duplicate key value") && strings.Contains(errMsg, "uq_employee_number") {
		return employeeerrors.ErrEmployeeNumberAlreadyExists
	}
	if strings.Contains(errMsg, "duplicate key value") && strings.Contains(errMsg, "uq_employee_email") {
		return employeeerrors.ErrEmployeeAlreadyExists
	}

	return err
}
