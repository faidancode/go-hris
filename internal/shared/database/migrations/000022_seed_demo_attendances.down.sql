DELETE FROM attendances
WHERE employee_id IN (
    SELECT e.id
    FROM employees e
    JOIN companies c ON c.id = e.company_id
    WHERE c.registration_number = 'DEMO-001'
      AND e.email IN ('owner@demo.com', 'hr@demo.com', 'employee@demo.com')
);
