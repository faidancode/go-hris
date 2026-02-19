DROP INDEX IF EXISTS idx_payrolls_payslip_generated_at;

ALTER TABLE payrolls
    DROP COLUMN IF EXISTS payslip_generated_at,
    DROP COLUMN IF EXISTS payslip_url;
