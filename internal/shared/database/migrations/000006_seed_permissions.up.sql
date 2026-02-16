-- =========================================
-- SIMPLE PERMISSION SEED (Portfolio Ready)
-- Idempotent
-- =========================================
INSERT INTO
    permissions (id, resource, action)
VALUES
    -- EMPLOYEE
    (gen_random_uuid (), 'employee', 'read'),
    (gen_random_uuid (), 'employee', 'create'),
    (gen_random_uuid (), 'employee', 'update'),
    (gen_random_uuid (), 'employee', 'delete'),
    -- DEPARTMENT
    (gen_random_uuid (), 'department', 'read'),
    (gen_random_uuid (), 'department', 'create'),
    (gen_random_uuid (), 'department', 'update'),
    (gen_random_uuid (), 'department', 'delete'),
    -- SALARY
    (gen_random_uuid (), 'salary', 'read'),
    (gen_random_uuid (), 'salary', 'update'),
    -- PAYROLL
    (gen_random_uuid (), 'payroll', 'read'),
    (gen_random_uuid (), 'payroll', 'approve'),
    -- ROLE MANAGEMENT
    (gen_random_uuid (), 'role', 'read'),
    (gen_random_uuid (), 'role', 'manage') ON CONFLICT (resource, action) DO NOTHING;