DROP INDEX IF EXISTS idx_payroll_components_company_type;
DROP INDEX IF EXISTS idx_payroll_components_payroll_id;
DROP TABLE IF EXISTS payroll_components;

ALTER TABLE payrolls
    DROP COLUMN IF EXISTS overtime_amount,
    DROP COLUMN IF EXISTS overtime_rate,
    DROP COLUMN IF EXISTS overtime_hours;
