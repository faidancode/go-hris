CREATE TABLE company_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL
        REFERENCES companies(id)
        ON DELETE CASCADE,
    type registration_type NOT NULL,
    number VARCHAR(100) NOT NULL,
    issued_at DATE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),

    CONSTRAINT uq_company_registration
        UNIQUE (company_id, type)
);

CREATE INDEX idx_company_registrations_company_id
ON company_registrations(company_id);

CREATE INDEX idx_company_registrations_number
ON company_registrations(number);