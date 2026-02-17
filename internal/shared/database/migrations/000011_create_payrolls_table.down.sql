-- Hapus index
DROP INDEX IF EXISTS idx_payrolls_created_at;
DROP INDEX IF EXISTS idx_payrolls_employee_period;
DROP INDEX IF EXISTS idx_payrolls_company_status;

-- Hapus tabel
DROP TABLE IF EXISTS payrolls;