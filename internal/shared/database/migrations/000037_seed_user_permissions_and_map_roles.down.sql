-- Remove role mappings for user module permissions.
DELETE FROM role_permissions rp
USING roles r, permissions p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND p.resource = 'user'
  AND p.action IN ('read', 'create', 'update')
  AND UPPER(r.name) IN ('SUPERADMIN', 'ADMIN', 'HR', 'OWNER');

-- Remove user module permissions.
DELETE FROM permissions
WHERE resource = 'user'
  AND action IN ('read', 'create', 'update');
