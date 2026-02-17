ALTER TABLE employees
DROP CONSTRAINT IF EXISTS fk_employees_position;

ALTER TABLE employees
DROP COLUMN IF EXISTS position_id;

DROP TABLE IF EXISTS positions;