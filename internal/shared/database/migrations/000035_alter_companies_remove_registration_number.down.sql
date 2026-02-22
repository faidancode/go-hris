ALTER TABLE companies
ADD COLUMN registration_number VARCHAR(100);

ALTER TABLE companies
ADD CONSTRAINT uq_companies_registration
UNIQUE (registration_number);