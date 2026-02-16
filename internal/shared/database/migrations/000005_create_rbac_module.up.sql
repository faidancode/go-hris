-- =========================================
-- TABLE: roles
-- =========================================
CREATE TABLE
    roles (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        company_id UUID NOT NULL,
        name VARCHAR(100) NOT NULL,
        description VARCHAR(255),
        created_at TIMESTAMP NOT NULL DEFAULT now (),
        updated_at TIMESTAMP NOT NULL DEFAULT now (),
        CONSTRAINT fk_roles_company FOREIGN KEY (company_id) REFERENCES companies (id) ON DELETE CASCADE,
        CONSTRAINT uq_roles_company_name UNIQUE (company_id, name)
    );

CREATE INDEX idx_roles_company_id ON roles (company_id);

-- =========================================
-- TABLE: permissions
-- Global permission definition
-- =========================================
CREATE TABLE
    permissions (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        resource VARCHAR(100) NOT NULL,
        action VARCHAR(50) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT now (),
        CONSTRAINT uq_permissions_resource_action UNIQUE (resource, action)
    );

CREATE INDEX idx_permissions_resource ON permissions (resource);

CREATE INDEX idx_permissions_action ON permissions (action);

-- =========================================
-- TABLE: role_permissions
-- =========================================
CREATE TABLE
    role_permissions (
        role_id UUID NOT NULL,
        permission_id UUID NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT now (),
        PRIMARY KEY (role_id, permission_id),
        CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
        CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
    );

CREATE INDEX idx_role_permissions_role_id ON role_permissions (role_id);

CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);

-- =========================================
-- TABLE: employee_roles
-- =========================================
CREATE TABLE
    employee_roles (
        employee_id UUID NOT NULL,
        role_id UUID NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT now (),
        PRIMARY KEY (employee_id, role_id),
        CONSTRAINT fk_employee_roles_employee FOREIGN KEY (employee_id) REFERENCES employees (id) ON DELETE CASCADE,
        CONSTRAINT fk_employee_roles_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
    );

CREATE INDEX idx_employee_roles_employee_id ON employee_roles (employee_id);

CREATE INDEX idx_employee_roles_role_id ON employee_roles (role_id);