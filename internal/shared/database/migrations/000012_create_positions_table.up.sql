-- 1. Buat Tabel Positions dengan link ke Department
CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    department_id UUID NOT NULL, -- Relasi ke Department
    name VARCHAR(120) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT fk_positions_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE,
    CONSTRAINT fk_positions_department FOREIGN KEY (department_id) REFERENCES departments (id) ON DELETE CASCADE
);

-- 2. Indexing
CREATE INDEX idx_positions_company_id ON positions (company_id);
CREATE INDEX idx_positions_department_id ON positions (department_id);
CREATE INDEX idx_positions_deleted_at ON positions (deleted_at);

-- 3. Update Employees (Hanya tambah position_id karena department_id biasanya sudah ada)
ALTER TABLE employees 
ADD COLUMN IF NOT EXISTS position_id UUID;

ALTER TABLE employees
ADD CONSTRAINT fk_employees_position 
FOREIGN KEY (position_id) REFERENCES positions (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_employees_position_id ON employees (position_id);