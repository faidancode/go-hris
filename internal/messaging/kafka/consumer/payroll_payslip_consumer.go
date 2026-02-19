package consumer

import (
	"context"
	"encoding/json"
	"go-hris/internal/events"
	"go-hris/internal/payroll"

	kafkago "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func ConsumePayrollPayslipRequested(
	ctx context.Context,
	reader *kafkago.Reader,
	payrollService payroll.Service,
	logger *zap.Logger,
) {
	log := logger.Named("kafka.consumer.payroll_payslip")
	log.Info("payroll payslip consumer started")

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Info("payroll payslip consumer stopped")
				return
			}
			log.Error("fetch payroll payslip message failed", zap.Error(err))
			continue
		}

		var event events.PayrollPayslipRequestedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Error("decode payroll payslip event failed", zap.Error(err))
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		_, err = payrollService.GeneratePayslip(ctx, event.CompanyID, event.PayrollID)
		if err != nil {
			log.Error("generate payslip failed",
				zap.String("payroll_id", event.PayrollID),
				zap.String("company_id", event.CompanyID),
				zap.Error(err),
			)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Error("commit payroll payslip message failed", zap.Error(err))
			continue
		}

		log.Info("payroll payslip generated",
			zap.String("payroll_id", event.PayrollID),
			zap.String("company_id", event.CompanyID),
		)
	}
}
