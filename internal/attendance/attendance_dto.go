package attendance

type ClockInRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Source    string   `json:"source"`
	Notes     *string  `json:"notes"`
}

type ClockOutRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	Notes     *string  `json:"notes"`
}

type AttendanceResponse struct {
	ID             string   `json:"id"`
	CompanyID      string   `json:"company_id"`
	EmployeeID     string   `json:"employee_id"`
	EmployeeName   string   `json:"employee_name,omitempty"`
	AttendanceDate string   `json:"attendance_date"`
	ClockIn        string   `json:"clock_in"`
	ClockOut       *string  `json:"clock_out,omitempty"`
	Latitude       *float64 `json:"latitude,omitempty"`
	Longitude      *float64 `json:"longitude,omitempty"`
	Status         string   `json:"status"`
	Source         string   `json:"source"`
	ExternalRef    *string  `json:"external_ref,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
}
