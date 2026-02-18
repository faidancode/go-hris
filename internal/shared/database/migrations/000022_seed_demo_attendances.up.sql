-- Demo attendance data for portfolio scenarios.
INSERT INTO attendances (
    id,
    company_id,
    employee_id,
    attendance_date,
    clock_in,
    clock_out,
    status,
    source,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid(),
    e.company_id,
    e.id,
    CURRENT_DATE,
    date_trunc('day', now()) + interval '08 hours 45 minutes',
    date_trunc('day', now()) + interval '17 hours 15 minutes',
    'PRESENT',
    'MANUAL',
    now(),
    now()
FROM employees e
JOIN companies c ON c.id = e.company_id
WHERE c.registration_number = 'DEMO-001'
  AND e.email IN ('owner@demo.com', 'hr@demo.com', 'employee@demo.com')
  AND e.deleted_at IS NULL
ON CONFLICT (employee_id, attendance_date) DO NOTHING;
