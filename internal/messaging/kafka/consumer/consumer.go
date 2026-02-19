package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"go-hris/internal/employeesalary"
	"go-hris/internal/events"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func ConsumeEmployeeLifecycle(
	ctx context.Context,
	reader *kafkago.Reader,
	employeeSalaryService employeesalary.Service,
	logger *zap.Logger,
) {
	log := logger.Named("kafka.consumer.employee_lifecycle")
	log.Info("employee lifecycle consumer started")

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Info("employee lifecycle consumer stopped")
				return
			}
			log.Error("fetch employee lifecycle message failed", zap.Error(err))
			continue
		}

		var event events.EmployeeCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Error("decode employee_created event failed", zap.Error(err))
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		effectiveDate := time.Now().UTC().Format("2006-01-02")
		_, err = employeeSalaryService.Create(ctx, event.CompanyID, employeesalary.CreateEmployeeSalaryRequest{
			EmployeeID:    event.EmployeeID,
			BaseSalary:    0,
			EffectiveDate: effectiveDate,
		})
		if err != nil {
			if isUniqueSalaryViolation(err) {
				log.Warn("employee salary already exists for event, skipping",
					zap.String("employee_id", event.EmployeeID),
					zap.String("company_id", event.CompanyID),
				)
				_ = reader.CommitMessages(ctx, msg)
				continue
			}

			log.Error("create default employee salary failed",
				zap.String("employee_id", event.EmployeeID),
				zap.String("company_id", event.CompanyID),
				zap.Error(err),
			)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Error("commit employee lifecycle message failed", zap.Error(err))
			continue
		}

		log.Info("employee salary created from employee_created event",
			zap.String("employee_id", event.EmployeeID),
			zap.String("company_id", event.CompanyID),
		)
	}
}

func isUniqueSalaryViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == "uq_employee_salary_effective"
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate key value") && strings.Contains(errMsg, "uq_employee_salary_effective")
}
