-- =========================================
-- ASSIGN OWNER ROLE
-- =========================================
INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT
    e.id,
    r.id,
    now()
FROM employees e
JOIN companies c
    ON c.id = e.company_id
JOIN roles r
    ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND e.email = 'owner@demo.com'
  AND r.name = 'Owner'
ON CONFLICT DO NOTHING;



-- =========================================
-- ASSIGN HR ROLE
-- =========================================
INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT
    e.id,
    r.id,
    now()
FROM employees e
JOIN companies c
    ON c.id = e.company_id
JOIN roles r
    ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND e.email = 'hr@demo.com'
  AND r.name = 'HR'
ON CONFLICT DO NOTHING;



-- =========================================
-- ASSIGN EMPLOYEE ROLE
-- =========================================
INSERT INTO employee_roles (employee_id, role_id, created_at)
SELECT
    e.id,
    r.id,
    now()
FROM employees e
JOIN companies c
    ON c.id = e.company_id
JOIN roles r
    ON r.company_id = c.id
WHERE c.registration_number = 'DEMO-001'
  AND e.email = 'employee@demo.com'
  AND r.name = 'Employee'
ON CONFLICT DO NOTHING;
