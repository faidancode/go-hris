-- 1. Remove permissions for Admin role
DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name = 'Admin');

-- 2. Remove Admin role
DELETE FROM roles WHERE name = 'Admin';
