-- =========================================
-- Idempotent: ON CONFLICT DO NOTHING
-- =========================================
INSERT INTO
    permissions (id, resource, action)
VALUES
    -- EMPLOYEE (Full CRUD)
    (gen_random_uuid (), 'employee', 'read'),
    (gen_random_uuid (), 'employee', 'create'),
    (gen_random_uuid (), 'employee', 'update'),
    (gen_random_uuid (), 'employee', 'delete'),
    -- DEPARTMENT & POSITION (Organizational Structure)
    (gen_random_uuid (), 'department', 'read'),
    (gen_random_uuid (), 'department', 'create'),
    (gen_random_uuid (), 'department', 'update'),
    (gen_random_uuid (), 'department', 'delete'),
    (gen_random_uuid (), 'position', 'read'),
    (gen_random_uuid (), 'position', 'create'),
    (gen_random_uuid (), 'position', 'update'),
    (gen_random_uuid (), 'position', 'delete'),
    -- SALARY (Value Management)
    (gen_random_uuid (), 'salary', 'read'),
    (gen_random_uuid (), 'salary', 'update'),
    -- PAYROLL (Workflow Based)
    (gen_random_uuid (), 'payroll', 'read'),
    (gen_random_uuid (), 'payroll', 'create'), -- Menjalankan kalkulasi/generate
    (gen_random_uuid (), 'payroll', 'approve'), -- Verifikasi angka
    (gen_random_uuid (), 'payroll', 'pay'), -- Finalisasi pembayaran
    (gen_random_uuid (), 'payroll', 'delete'), -- Batalkan draft jika salah
    -- LEAVE
    (gen_random_uuid (), 'leave', 'read'), -- Melihat daftar cuti
    (gen_random_uuid (), 'leave', 'create'), -- Mengajukan cuti
    (gen_random_uuid (), 'leave', 'approve'), -- Menyetujui/Menolak cuti
    (gen_random_uuid (), 'leave', 'manage') -- Mengatur kuota cuti tahunan
    -- ROLE MANAGEMENT
    (gen_random_uuid (), 'role', 'read'),
    (gen_random_uuid (), 'role', 'manage'),
    -- COMPANY SETTINGS
    (gen_random_uuid (), 'company', 'read'),
    (gen_random_uuid (), 'company', 'update'),
    -- ATTENDANCE
    (gen_random_uuid (), 'attendance', 'read'),
    (gen_random_uuid (), 'attendance', 'manage') ON CONFLICT (resource, action) DO NOTHING;