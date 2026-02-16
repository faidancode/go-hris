DELETE FROM employees
WHERE email IN (
    'owner@demo.com',
    'hr@demo.com',
    'employee@demo.com'
);

DELETE FROM companies
WHERE registration_number = 'DEMO-001';
