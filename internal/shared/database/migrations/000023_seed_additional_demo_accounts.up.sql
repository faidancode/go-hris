-- Seed additional demo accounts for QA/testing:
-- - finance@demo.com  -> Finance role
-- - employee2@demo.com -> Employee role
-- - employee3@demo.com -> Employee role
-- Default plain password: password123

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1) Ensure demo employees exist (idempotent by company_id+email).
INSERT INTO employees (
    id,
    company_id,
    employee_number,
    full_name,
    email,
    hire_date,
    employment_status,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid(),
    c.id,
    seed.employee_number,
    seed.full_name,
    seed.email,
    CURRENT_DATE,
    'active',
    now(),
    now()
FROM companies c
JOIN (
    VALUES
        ('EMP-901', 'Finance User', 'finance@demo.com'),
        ('EMP-902', 'Employee 2 User', 'employee2@demo.com'),
        ('EMP-903', 'Employee 3 User', 'employee3@demo.com')
) AS seed(employee_number, full_name, email) ON true
WHERE c.registration_number = 'DEMO-001'
ON CONFLICT (company_id, email) DO UPDATE
SET
    full_name = EXCLUDED.full_name,
    updated_at = now();

-- 2) Upsert users so login always works with predictable credentials.
INSERT INTO users (
    id,
    employee_id,
    company_id,
    name,
    email,
    password,
    role,
    is_active,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    gen_random_uuid(),
    e.id,
    e.company_id,
    e.full_name,
    e.email,
    crypt('password123', gen_salt('bf')),
    CASE
        WHEN e.email = 'finance@demo.com' THEN 'FINANCE'
        ELSE 'EMPLOYEE'
    END,
    TRUE,
    now(),
    now(),
    NULL
FROM employees e
JOIN companies c ON c.id = e.company_id
WHERE c.registration_number = 'DEMO-001'
  AND e.email IN ('finance@demo.com', 'employee2@demo.com', 'employee3@demo.com')
  AND e.deleted_at IS NULL
ON CONFLICT (email) DO UPDATE
SET
    employee_id = EXCLUDED.employee_id,
    company_id = EXCLUDED.company_id,
    name = EXCLUDED.name,
    password = EXCLUDED.password,
    role = EXCLUDED.role,
    is_active = TRUE,
    updated_at = now(),
    deleted_at = NULL;

-- 3) Reset role assignments for these 3 users, then assign intended roles.
DELETE FROM employee_roles er
USING employees e, companies c
WHERE er.employee_id = e.id
  AND e.company_id = c.id
  AND c.registration_number = 'DEMO-001'
  AND e.email IN ('finance@demo.com', 'employee2@demo.com', 'employee3@demo.com');

INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT e.id, r.id, now()
FROM employees e
JOIN companies c ON c.id = e.company_id
JOIN roles r ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND e.email = 'finance@demo.com'
  AND r.name = 'Finance'
ON CONFLICT DO NOTHING;

INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT e.id, r.id, now()
FROM employees e
JOIN companies c ON c.id = e.company_id
JOIN roles r ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND e.email IN ('employee2@demo.com', 'employee3@demo.com')
  AND r.name = 'Employee'
ON CONFLICT DO NOTHING;
