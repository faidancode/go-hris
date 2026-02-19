DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id FROM roles WHERE LOWER(name) = 'finance'
)
AND permission_id IN (
    SELECT id FROM permissions WHERE resource = 'employee' AND action = 'read'
);
