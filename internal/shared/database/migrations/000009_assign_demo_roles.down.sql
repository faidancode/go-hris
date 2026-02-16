DELETE FROM employee_roles
WHERE employee_id IN (
    SELECT id FROM employees
    WHERE email IN (
        'owner@demo.com',
        'hr@demo.com',
        'employee@demo.com'
    )
);
