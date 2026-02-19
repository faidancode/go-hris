ALTER TABLE payrolls
    ADD COLUMN IF NOT EXISTS payslip_url TEXT,
    ADD COLUMN IF NOT EXISTS payslip_generated_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_payrolls_payslip_generated_at ON payrolls (payslip_generated_at);
