DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id
    FROM permissions
    WHERE resource = 'attendance' AND action = 'create'
);

DELETE FROM permissions
WHERE resource = 'attendance' AND action = 'create';
