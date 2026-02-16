-- =========================
-- Table: users
-- =========================
CREATE TABLE
    users (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        employee_id UUID NOT NULL REFERENCES employees (id) ON DELETE CASCADE,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL, -- bcrypt/argon2 hash
        is_active BOOLEAN DEFAULT TRUE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now ()
    );

-- Index untuk lookup cepat berdasarkan employee_id
CREATE INDEX idx_users_employee_id ON users (employee_id);

-- Index untuk lookup email saat login
CREATE INDEX idx_users_email ON users (email);