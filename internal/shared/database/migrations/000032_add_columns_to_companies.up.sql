-- Add email and is_active columns to companies table
ALTER TABLE companies ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE companies ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;

-- Ensure email is unique per company
CREATE UNIQUE INDEX IF NOT EXISTS uq_companies_email ON companies(email) WHERE deleted_at IS NULL;
