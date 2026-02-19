package employeesalary

import (
	"context"
	"encoding/json"
	"errors"
	"go-hris/internal/events"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type EmployeeCreatedConsumer struct {
	reader  *kafka.Reader
	service Service
	logger  *zap.Logger
}

func NewEmployeeCreatedConsumer(
	broker string,
	groupID string,
	service Service,
	logger ...*zap.Logger,
) *EmployeeCreatedConsumer {
	l := zap.L().Named("employeesalary.consumer")
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0].Named("employeesalary.consumer")
	}

	return &EmployeeCreatedConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        []string{broker},
			Topic:          events.EmployeeCreatedTopic,
			GroupID:        groupID,
			CommitInterval: time.Second,
			StartOffset:    kafka.FirstOffset,
		}),
		service: service,
		logger:  l,
	}
}

func (c *EmployeeCreatedConsumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.logger.Error("consume employee_created failed", zap.Error(err))
				continue
			}

			var event events.EmployeeCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				c.logger.Error("decode employee_created event failed", zap.Error(err))
				if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
					c.logger.Error("commit invalid employee_created event failed", zap.Error(commitErr))
				}
				continue
			}

			effectiveDate := time.Now().UTC().Format("2006-01-02")
			_, err = c.service.Create(ctx, event.CompanyID, CreateEmployeeSalaryRequest{
				EmployeeID:    event.EmployeeID,
				BaseSalary:    0,
				EffectiveDate: effectiveDate,
			})
			if err != nil {
				// Duplicate event is safe to skip.
				if isUniqueSalaryViolation(err) {
					c.logger.Warn("employee salary already exists for event, skipping",
						zap.String("employee_id", event.EmployeeID),
						zap.String("company_id", event.CompanyID),
					)
					if commitErr := c.reader.CommitMessages(ctx, msg); commitErr != nil {
						c.logger.Error("commit duplicate employee_created event failed", zap.Error(commitErr))
					}
					continue
				}

				c.logger.Error("create default employee salary failed",
					zap.String("employee_id", event.EmployeeID),
					zap.String("company_id", event.CompanyID),
					zap.Error(err),
				)
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("commit employee_created event failed", zap.Error(err))
				continue
			}

			c.logger.Info("employee salary created from employee_created event",
				zap.String("employee_id", event.EmployeeID),
				zap.String("company_id", event.CompanyID),
			)
		}
	}()
}

func isUniqueSalaryViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == "uq_employee_salary_effective"
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "duplicate key value") && strings.Contains(errMsg, "uq_employee_salary_effective")
}
