CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    registration_number VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP
);

ALTER TABLE companies
ADD CONSTRAINT uq_companies_registration UNIQUE (registration_number);

CREATE INDEX idx_companies_deleted_at ON companies(deleted_at);