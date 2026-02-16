-- Remove role_permissions first
DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id FROM roles
    WHERE name IN ('Owner', 'HR', 'Employee')
);

-- Remove roles
DELETE FROM roles
WHERE name IN ('Owner', 'HR', 'Employee');
