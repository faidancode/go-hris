package events

import "time"

const EmployeeCreatedTopic = "hr.employee.lifecycle.v1"

type EmployeeCreatedEvent struct {
	RequestID  string    `json:"request_id"`
	EventType  string    `json:"event_type"`
	EmployeeID string    `json:"employee_id"`
	CompanyID  string    `json:"company_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
