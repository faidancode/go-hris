-- Seed demo users from seeded demo employees.
-- Default plain password for demo login: password123
-- Stored as bcrypt hash via pgcrypto: crypt(..., gen_salt('bf'))

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DO $$
DECLARE
    has_company_id BOOLEAN;
    has_name BOOLEAN;
    has_role BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'company_id'
    ) INTO has_company_id;

    SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'name'
    ) INTO has_name;

    SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'role'
    ) INTO has_role;

    IF has_company_id AND has_name AND has_role THEN
        INSERT INTO users (
            id,
            employee_id,
            company_id,
            name,
            email,
            password,
            role,
            is_active,
            created_at,
            updated_at
        )
        SELECT
            gen_random_uuid(),
            e.id,
            e.company_id,
            e.full_name,
            e.email,
            crypt('password123', gen_salt('bf')),
            CASE
                WHEN e.email = 'owner@demo.com' THEN 'OWNER'
                WHEN e.email = 'hr@demo.com' THEN 'HR'
                ELSE 'EMPLOYEE'
            END,
            TRUE,
            now(),
            now()
        FROM employees e
        WHERE e.email IN ('owner@demo.com', 'hr@demo.com', 'employee@demo.com')
          AND e.deleted_at IS NULL
        ON CONFLICT (email) DO NOTHING;
    ELSE
        INSERT INTO users (
            id,
            employee_id,
            email,
            password,
            is_active,
            created_at,
            updated_at
        )
        SELECT
            gen_random_uuid(),
            e.id,
            e.email,
            crypt('password123', gen_salt('bf')),
            TRUE,
            now(),
            now()
        FROM employees e
        WHERE e.email IN ('owner@demo.com', 'hr@demo.com', 'employee@demo.com')
          AND e.deleted_at IS NULL
        ON CONFLICT (email) DO NOTHING;
    END IF;
END $$;
