DELETE FROM permissions
WHERE
    (resource, action) IN (
        -- Employee
        ('employee', 'read'),
        ('employee', 'create'),
        ('employee', 'update'),
        ('employee', 'delete'),
        -- Department and Position
        ('department', 'read'),
        ('department', 'create'),
        ('department', 'update'),
        ('department', 'delete'),
        ('position', 'read'),
        ('position', 'create'),
        ('position', 'update'),
        ('position', 'delete'),
        -- Salary
        ('salary', 'read'),
        ('salary', 'update'),
        -- Payroll
        ('payroll', 'read'),
        ('payroll', 'create'),
        ('payroll', 'approve'),
        ('payroll', 'pay'),
        ('payroll', 'delete'),
        -- Leave
        ('leave', 'read'),
        ('leave', 'create'),
        ('leave', 'approve'),
        ('leave', 'manage')
        -- Role
        ('role', 'read'),
        ('role', 'manage'),
        -- Company
        ('company', 'read'),
        ('company', 'update'),
        -- Attendance
        ('attendance', 'read'),
        ('attendance', 'manage')
    );