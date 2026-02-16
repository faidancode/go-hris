CREATE TABLE
    employee_salaries (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        employee_id UUID NOT NULL,
        base_salary INT NOT NULL, -- stored in cents
        effective_date DATE NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT now (),
        updated_at TIMESTAMP NOT NULL DEFAULT now (),
        CONSTRAINT fk_salary_employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE
    );

ALTER TABLE employee_salaries ADD CONSTRAINT uq_employee_salary_effective UNIQUE (employee_id, effective_date);

CREATE INDEX idx_salary_employee_id ON employee_salaries (employee_id);

CREATE INDEX idx_salary_effective_date ON employee_salaries (effective_date DESC);