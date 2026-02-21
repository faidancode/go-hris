-- 1. Create Admin role for all existing companies
INSERT INTO roles (id, company_id, name, description, created_at, updated_at)
SELECT 
    gen_random_uuid(), 
    c.id, 
    'Admin', 
    'Administrator dengan akses penuh ke semua modul perusahaan', 
    now(), 
    now()
FROM companies c
ON CONFLICT (company_id, name) DO NOTHING;

-- 2. Grant all current permissions to Admin role
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, now()
FROM roles r
JOIN permissions p ON true
WHERE r.name = 'Admin'
ON CONFLICT DO NOTHING;
