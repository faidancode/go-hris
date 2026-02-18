CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS attendances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    attendance_date DATE NOT NULL,
    clock_in TIMESTAMPTZ NOT NULL,
    clock_out TIMESTAMPTZ,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    status VARCHAR(20) NOT NULL DEFAULT 'PRESENT',
    source VARCHAR(30) NOT NULL DEFAULT 'MANUAL',
    external_ref VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_attendances_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE,
    CONSTRAINT fk_attendances_employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE,
    CONSTRAINT uq_attendances_employee_date UNIQUE (employee_id, attendance_date),
    CONSTRAINT chk_attendances_status CHECK (status IN ('PRESENT', 'LATE', 'ABSENT')),
    CONSTRAINT chk_attendances_clock_range CHECK (clock_out IS NULL OR clock_out >= clock_in)
);

CREATE INDEX IF NOT EXISTS idx_attendances_company_date ON attendances (company_id, attendance_date);
CREATE INDEX IF NOT EXISTS idx_attendances_employee_date ON attendances (employee_id, attendance_date);
CREATE INDEX IF NOT EXISTS idx_attendances_status ON attendances (status);
CREATE INDEX IF NOT EXISTS idx_attendances_deleted_at ON attendances (deleted_at);
