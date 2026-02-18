-- Rollback bootstrap SUPERADMIN role mapping.

DELETE FROM employee_roles
WHERE role_id IN (
    SELECT id
    FROM roles
    WHERE name = 'SUPERADMIN'
);

DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id
    FROM roles
    WHERE name = 'SUPERADMIN'
);

DELETE FROM roles
WHERE name = 'SUPERADMIN';
