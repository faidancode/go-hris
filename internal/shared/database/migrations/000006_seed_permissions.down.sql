DELETE FROM permissions
WHERE (resource, action) IN (
    ('employee', 'read'),
    ('employee', 'create'),
    ('employee', 'update'),
    ('employee', 'delete'),

    ('department', 'read'),
    ('department', 'create'),
    ('department', 'update'),
    ('department', 'delete'),

    ('salary', 'read'),
    ('salary', 'update'),

    ('payroll', 'read'),
    ('payroll', 'approve'),

    ('role', 'read'),
    ('role', 'manage')
);
