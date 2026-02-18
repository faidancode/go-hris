-- Align users table with auth.User entity used by GORM (including soft delete).

ALTER TABLE users
ADD COLUMN IF NOT EXISTS company_id UUID,
ADD COLUMN IF NOT EXISTS name VARCHAR(255),
ADD COLUMN IF NOT EXISTS role VARCHAR(50),
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Backfill company_id and name from employees.
UPDATE users u
SET
    company_id = e.company_id,
    name = COALESCE(u.name, e.full_name)
FROM employees e
WHERE u.employee_id = e.id
  AND (u.company_id IS NULL OR u.name IS NULL);

-- Backfill role default for existing rows.
UPDATE users
SET role = 'EMPLOYEE'
WHERE role IS NULL OR role = '';

ALTER TABLE users
ALTER COLUMN role SET DEFAULT 'EMPLOYEE';

-- Add FK for company_id if missing.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_users_company'
    ) THEN
        ALTER TABLE users
        ADD CONSTRAINT fk_users_company
        FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Apply NOT NULL when data is complete.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM users WHERE company_id IS NULL) THEN
        RAISE NOTICE 'users.company_id still has NULL values, skip SET NOT NULL';
    ELSE
        ALTER TABLE users ALTER COLUMN company_id SET NOT NULL;
    END IF;

    IF EXISTS (SELECT 1 FROM users WHERE name IS NULL) THEN
        RAISE NOTICE 'users.name still has NULL values, skip SET NOT NULL';
    ELSE
        ALTER TABLE users ALTER COLUMN name SET NOT NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_users_company_id ON users (company_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
