-- Ensure employees.position_id exists and linked to positions.
-- Idempotent migration for environments that missed previous schema changes.

ALTER TABLE employees
ADD COLUMN IF NOT EXISTS position_id UUID;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_employees_position'
    ) THEN
        ALTER TABLE employees
        ADD CONSTRAINT fk_employees_position
        FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_employees_position_id ON employees(position_id);
