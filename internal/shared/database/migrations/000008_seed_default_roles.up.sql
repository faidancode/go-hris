-- =========================================
-- SEED DEFAULT ROLES PER COMPANY
-- Uses gen_random_uuid() (Postgres 18)
-- Idempotent
-- =========================================
INSERT INTO
    roles (
        id,
        company_id,
        name,
        description,
        created_at,
        updated_at
    )
SELECT
    gen_random_uuid (),
    c.id,
    r.name,
    r.description,
    now (),
    now ()
FROM
    companies c
    CROSS JOIN (
        VALUES
            ('Owner', 'Akses penuh ke semua modul'),
            (
                'HR',
                'Mengelola karyawan, cuti, dan draft payroll'
            ),
            (
                'Finance',
                'Mengelola gaji, persetujuan payroll, dan pembayaran'
            ),
            (
                'Employee',
                'Hanya akses baca profil, slip gaji, dan pengajuan cuti'
            )
    ) AS r (name, description) ON CONFLICT (company_id, name) DO NOTHING;

-- 3️⃣ MAPPING PERMISSIONS TO ROLES
-- OWNER: Full Power
INSERT INTO
    role_permissions (role_id, permission_id, created_at)
SELECT
    ro.id,
    p.id,
    now ()
FROM
    roles ro
    JOIN permissions p ON true
WHERE
    ro.name = 'Owner' ON CONFLICT DO NOTHING;

-- HR: Org Structure, Employees, Attendance, Leave (Approve), Payroll (Create)
INSERT INTO
    role_permissions (role_id, permission_id, created_at)
SELECT
    ro.id,
    p.id,
    now ()
FROM
    roles ro
    JOIN permissions p ON (
        p.resource IN (
            'employee',
            'department',
            'position',
            'attendance'
        )
    )
    OR (
        p.resource = 'leave'
        AND p.action IN ('read', 'approve', 'manage')
    )
    OR (
        p.resource = 'payroll'
        AND p.action IN ('read', 'create', 'delete')
    )
WHERE
    ro.name = 'HR' ON CONFLICT DO NOTHING;

-- FINANCE: Salary & Payroll (Finalization)
INSERT INTO
    role_permissions (role_id, permission_id, created_at)
SELECT
    ro.id,
    p.id,
    now ()
FROM
    roles ro
    JOIN permissions p ON (p.resource = 'salary')
    OR (
        p.resource = 'payroll'
        AND p.action IN ('read', 'approve', 'pay')
    )
WHERE
    ro.name = 'Finance' ON CONFLICT DO NOTHING;

-- EMPLOYEE: Self-Service Access
INSERT INTO
    role_permissions (role_id, permission_id, created_at)
SELECT
    ro.id,
    p.id,
    now ()
FROM
    roles ro
    JOIN permissions p ON (
        p.resource = 'employee'
        AND p.action = 'read'
    )
    OR (
        p.resource = 'payroll'
        AND p.action = 'read'
    )
    OR (
        p.resource = 'leave'
        AND p.action IN ('read', 'create')
    )
WHERE
    ro.name = 'Employee' ON CONFLICT DO NOTHING;