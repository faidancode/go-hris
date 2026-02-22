ALTER TABLE companies
DROP CONSTRAINT IF EXISTS uq_companies_registration;

ALTER TABLE companies
DROP COLUMN IF EXISTS registration_number;