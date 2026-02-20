ALTER TABLE permissions ADD COLUMN label VARCHAR(100);
ALTER TABLE permissions ADD COLUMN category VARCHAR(100);

-- Seed metadata for existing permissions
UPDATE permissions SET label = 'Melihat Karyawan', category = 'Karyawan' WHERE resource = 'employee' AND action = 'read';
UPDATE permissions SET label = 'Tambah Karyawan', category = 'Karyawan' WHERE resource = 'employee' AND action = 'create';
UPDATE permissions SET label = 'Edit Karyawan', category = 'Karyawan' WHERE resource = 'employee' AND action = 'update';
UPDATE permissions SET label = 'Hapus Karyawan', category = 'Karyawan' WHERE resource = 'employee' AND action = 'delete';

UPDATE permissions SET label = 'Melihat Departemen', category = 'Organisasi' WHERE resource = 'department' AND action = 'read';
UPDATE permissions SET label = 'Tambah Departemen', category = 'Organisasi' WHERE resource = 'department' AND action = 'create';
UPDATE permissions SET label = 'Edit Departemen', category = 'Organisasi' WHERE resource = 'department' AND action = 'update';
UPDATE permissions SET label = 'Hapus Departemen', category = 'Organisasi' WHERE resource = 'department' AND action = 'delete';

UPDATE permissions SET label = 'Melihat Jabatan', category = 'Organisasi' WHERE resource = 'position' AND action = 'read';
UPDATE permissions SET label = 'Tambah Jabatan', category = 'Organisasi' WHERE resource = 'position' AND action = 'create';
UPDATE permissions SET label = 'Edit Jabatan', category = 'Organisasi' WHERE resource = 'position' AND action = 'update';
UPDATE permissions SET label = 'Hapus Jabatan', category = 'Organisasi' WHERE resource = 'position' AND action = 'delete';

UPDATE permissions SET label = 'Melihat Gaji', category = 'Gaji' WHERE resource = 'salary' AND action = 'read';
UPDATE permissions SET label = 'Edit Gaji', category = 'Gaji' WHERE resource = 'salary' AND action = 'update';

UPDATE permissions SET label = 'Melihat Payroll', category = 'Payroll' WHERE resource = 'payroll' AND action = 'read';
UPDATE permissions SET label = 'Generate Payroll', category = 'Payroll' WHERE resource = 'payroll' AND action = 'create';
UPDATE permissions SET label = 'Setujui Payroll', category = 'Payroll' WHERE resource = 'payroll' AND action = 'approve';
UPDATE permissions SET label = 'Bayar Payroll', category = 'Payroll' WHERE resource = 'payroll' AND action = 'pay';
UPDATE permissions SET label = 'Hapus Payroll', category = 'Payroll' WHERE resource = 'payroll' AND action = 'delete';

UPDATE permissions SET label = 'Melihat Cuti', category = 'Cuti' WHERE resource = 'leave' AND action = 'read';
UPDATE permissions SET label = 'Ajukan Cuti', category = 'Cuti' WHERE resource = 'leave' AND action = 'create';
UPDATE permissions SET label = 'Setujui Cuti', category = 'Cuti' WHERE resource = 'leave' AND action = 'approve';
UPDATE permissions SET label = 'Kelola Kuota Cuti', category = 'Cuti' WHERE resource = 'leave' AND action = 'manage';

UPDATE permissions SET label = 'Melihat Role', category = 'User Management' WHERE resource = 'role' AND action = 'read';
UPDATE permissions SET label = 'Kelola Role & Permission', category = 'User Management' WHERE resource = 'role' AND action = 'manage';

UPDATE permissions SET label = 'Melihat Profil Perusahaan', category = 'Pengaturan' WHERE resource = 'company' AND action = 'read';
UPDATE permissions SET label = 'Edit Profil Perusahaan', category = 'Pengaturan' WHERE resource = 'company' AND action = 'update';

UPDATE permissions SET label = 'Melihat Presensi', category = 'Presensi' WHERE resource = 'attendance' AND action = 'read';
UPDATE permissions SET label = 'Kelola Presensi', category = 'Presensi' WHERE resource = 'attendance' AND action = 'manage';
