CREATE TABLE
    departments (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        company_id UUID NOT NULL,
        name VARCHAR(120) NOT NULL,
        parent_department_id UUID,
        created_at TIMESTAMP NOT NULL DEFAULT now (),
        updated_at TIMESTAMP NOT NULL DEFAULT now (),
        deleted_at TIMESTAMP,
        CONSTRAINT fk_departments_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE,
        CONSTRAINT fk_departments_parent FOREIGN KEY (parent_department_id) REFERENCES departments (id) ON DELETE SET NULL
    );

CREATE INDEX idx_departments_company_id ON departments (company_id);

CREATE INDEX idx_departments_parent_id ON departments (parent_department_id);

CREATE INDEX idx_departments_deleted_at ON departments (deleted_at);