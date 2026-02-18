-- Rollback additional demo accounts seed.

DELETE FROM employee_roles er
USING employees e, companies c
WHERE er.employee_id = e.id
  AND e.company_id = c.id
  AND c.registration_number = 'DEMO-001'
  AND e.email IN ('finance@demo.com', 'employee2@demo.com', 'employee3@demo.com');

DELETE FROM users
WHERE email IN ('finance@demo.com', 'employee2@demo.com', 'employee3@demo.com');

DELETE FROM employees e
USING companies c
WHERE e.company_id = c.id
  AND c.registration_number = 'DEMO-001'
  AND e.email IN ('finance@demo.com', 'employee2@demo.com', 'employee3@demo.com');
