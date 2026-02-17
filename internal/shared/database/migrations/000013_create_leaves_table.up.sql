CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS leaves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    employee_id UUID NOT NULL,

    leave_type VARCHAR(30) NOT NULL DEFAULT 'ANNUAL',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    total_days INT NOT NULL DEFAULT 1,
    reason TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_by UUID NOT NULL,
    approved_by UUID,
    rejection_reason TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_leaves_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE,
    CONSTRAINT fk_leaves_employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE,
    CONSTRAINT fk_leaves_creator FOREIGN KEY (created_by) REFERENCES employees (id) ON DELETE RESTRICT,
    CONSTRAINT fk_leaves_approver FOREIGN KEY (approved_by) REFERENCES employees (id) ON DELETE RESTRICT,
    CONSTRAINT chk_leaves_date_range CHECK (start_date <= end_date),
    CONSTRAINT chk_leaves_total_days CHECK (total_days > 0)
);

CREATE INDEX idx_leaves_company_status ON leaves (company_id, status);
CREATE INDEX idx_leaves_employee_dates ON leaves (employee_id, start_date, end_date);
CREATE INDEX idx_leaves_start_date ON leaves (start_date);
CREATE INDEX idx_leaves_deleted_at ON leaves (deleted_at);
