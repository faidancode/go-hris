package company

import "time"

type CompanyResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	RegistrationNumber string `json:"registration_number"`
	IsActive           bool   `json:"is_active"`
}

type UpdateCompanyRequest struct {
	Name               string `json:"name"`
	RegistrationNumber string `json:"registration_number"`
	IsActive           *bool  `json:"is_active"`
}

type UpsertCompanyRegistrationRequest struct {
	Type     RegistrationType `json:"type" binding:"required"`
	Number   string           `json:"number" binding:"required"`
	IssuedAt *time.Time       `json:"issued_at,omitempty"`
}

type CompanyRegistrationResponse struct {
	ID        string           `json:"id"`
	Type      RegistrationType `json:"type"`
	Number    string           `json:"number"`
	IssuedAt  *time.Time       `json:"issued_at,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
