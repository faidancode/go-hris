-- Pastikan ekstensi UUID tersedia (PostgreSQL)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE
    IF NOT EXISTS payrolls (
        -- Primary Key
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        -- Relasi Utama
        company_id UUID NOT NULL,
        employee_id UUID NOT NULL,
        -- Periode Penggajian
        period_start DATE NOT NULL,
        period_end DATE NOT NULL,
        -- Komponen Keuangan disimpan dalam satuan terkecil (mis: sen)
        base_salary BIGINT NOT NULL DEFAULT 0,
        allowance BIGINT NOT NULL DEFAULT 0,
        deduction BIGINT NOT NULL DEFAULT 0,
        net_salary BIGINT NOT NULL DEFAULT 0,
        -- Status Workflow
        -- Status: 'DRAFT', 'PENDING', 'APPROVED', 'PAID', 'CANCELLED'
        status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
        -- Audit Trail & Akuntabilitas
        created_by UUID NOT NULL,
        approved_by UUID,
        created_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            paid_at TIMESTAMP
        WITH
            TIME ZONE,
            approved_at TIMESTAMP
        WITH
            TIME ZONE,
            -- Foreign Key Constraints
            CONSTRAINT fk_payroll_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE,
            CONSTRAINT fk_payroll_employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE,
            CONSTRAINT fk_payroll_creator FOREIGN KEY (created_by) REFERENCES employees (id) ON DELETE RESTRICT,
            CONSTRAINT fk_payroll_approver FOREIGN KEY (approved_by) REFERENCES employees (id) ON DELETE RESTRICT,
            -- Business Logic: Satu karyawan hanya boleh punya satu record payroll per periode
            CONSTRAINT unique_employee_payroll_period UNIQUE (employee_id, period_start, period_end)
    );

-- Indexing untuk mempercepat filter dan pelaporan
CREATE INDEX idx_payrolls_company_status ON payrolls (company_id, status);

CREATE INDEX idx_payrolls_employee_period ON payrolls (employee_id, period_start, period_end);

CREATE INDEX idx_payrolls_created_at ON payrolls (created_at);
