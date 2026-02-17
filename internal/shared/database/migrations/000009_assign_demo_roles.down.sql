-- 1. Bersihkan Junction Table
DELETE FROM role_permissions;

-- 2. Bersihkan Roles
DELETE FROM roles
WHERE
    name IN ('Owner', 'HR', 'Finance', 'Employee');

-- 3. Bersihkan Permissions
DELETE FROM permissions
WHERE
    resource IN (
        'employee',
        'department',
        'position',
        'salary',
        'payroll',
        'leave',
        'role',
        'company',
        'attendance'
    );