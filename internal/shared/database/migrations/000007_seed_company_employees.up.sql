-- =========================================
-- SEED DEMO COMPANY
-- =========================================

INSERT INTO companies (
    id,
    name,
    registration_number,
    created_at,
    updated_at
)
VALUES (
    gen_random_uuid(),
    'Demo Company',
    'DEMO-001',
    now(),
    now()
)
ON CONFLICT (registration_number) DO NOTHING;



-- =========================================
-- SEED DEMO EMPLOYEES
-- =========================================

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
    e.employee_number,
    e.full_name,
    e.email,
    CURRENT_DATE,
    'active',
    now(),
    now()
FROM companies c
CROSS JOIN (
    VALUES
        ('EMP-001', 'Owner User', 'owner@demo.com'),
        ('EMP-002', 'HR User', 'hr@demo.com'),
        ('EMP-003', 'Employee User', 'employee@demo.com')
) AS e(employee_number, full_name, email)
WHERE c.registration_number = 'DEMO-001'
ON CONFLICT (company_id, employee_number) DO NOTHING;
