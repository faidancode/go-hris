package bootstrap

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type StdoutAuditLogger struct{}

func NewStdoutAuditLogger() *StdoutAuditLogger {
	return &StdoutAuditLogger{}
}

func (l *StdoutAuditLogger) Log(ctx context.Context, entry AuditLog) {
	zap.L().Named("audit").Info("audit event",
		zap.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
		zap.String("action", entry.Action),
		zap.String("message", entry.Message),
		zap.Any("meta", entry.Meta),
	)
}
