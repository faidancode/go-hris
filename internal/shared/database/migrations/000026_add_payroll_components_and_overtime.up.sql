ALTER TABLE payrolls
    ADD COLUMN IF NOT EXISTS overtime_hours BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS overtime_rate BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS overtime_amount BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS payroll_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payroll_id UUID NOT NULL,
    company_id UUID NOT NULL,
    component_type VARCHAR(20) NOT NULL,
    component_name VARCHAR(120) NOT NULL,
    quantity BIGINT NOT NULL DEFAULT 1,
    unit_amount BIGINT NOT NULL DEFAULT 0,
    total_amount BIGINT NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_payroll_components_payroll FOREIGN KEY (payroll_id) REFERENCES payrolls (id) ON DELETE CASCADE,
    CONSTRAINT fk_payroll_components_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE,
    CONSTRAINT chk_payroll_components_type CHECK (component_type IN ('ALLOWANCE', 'DEDUCTION')),
    CONSTRAINT chk_payroll_components_quantity CHECK (quantity > 0),
    CONSTRAINT chk_payroll_components_unit_amount CHECK (unit_amount >= 0),
    CONSTRAINT chk_payroll_components_total_amount CHECK (total_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_payroll_components_payroll_id ON payroll_components (payroll_id);
CREATE INDEX IF NOT EXISTS idx_payroll_components_company_type ON payroll_components (company_id, component_type);
