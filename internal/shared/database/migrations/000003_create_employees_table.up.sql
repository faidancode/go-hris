CREATE TABLE
    employees (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        company_id UUID NOT NULL,
        department_id UUID,
        employee_number VARCHAR(50) NOT NULL,
        full_name VARCHAR(150) NOT NULL,
        email VARCHAR(150) NOT NULL,
        phone VARCHAR(50),
        hire_date DATE NOT NULL,
        employment_status VARCHAR(30) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT now (),
        updated_at TIMESTAMP NOT NULL DEFAULT now (),
        deleted_at TIMESTAMP,
        CONSTRAINT fk_employees_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE,
        CONSTRAINT fk_employees_department FOREIGN KEY (department_id) REFERENCES departments (id) ON DELETE SET NULL
    );

ALTER TABLE employees ADD CONSTRAINT uq_employee_number UNIQUE (company_id, employee_number);

ALTER TABLE employees ADD CONSTRAINT uq_employee_email UNIQUE (company_id, email);

CREATE INDEX idx_employees_company_id ON employees (company_id);

CREATE INDEX idx_employees_department_id ON employees (department_id);

CREATE INDEX idx_employees_status ON employees (employment_status);

CREATE INDEX idx_employees_deleted_at ON employees (deleted_at);