ALTER TABLE companies DROP COLUMN IF EXISTS email;
ALTER TABLE companies DROP COLUMN IF EXISTS is_active;
DROP INDEX IF EXISTS uq_companies_email;
