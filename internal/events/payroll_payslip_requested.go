package events

import "time"

const PayrollPayslipRequestedTopic = "hr.payroll.payslip.requested.v1"

type PayrollPayslipRequestedEvent struct {
	EventType   string    `json:"event_type"`
	PayrollID   string    `json:"payroll_id"`
	CompanyID   string    `json:"company_id"`
	RequestedBy string    `json:"requested_by"`
	OccurredAt  time.Time `json:"occurred_at"`
}
