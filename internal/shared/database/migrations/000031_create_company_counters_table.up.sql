-- =========================================
-- TABLE: company_counters
-- To store atomic sequence/counters per company
-- =========================================
CREATE TABLE IF NOT EXISTS company_counters (
    company_id UUID NOT NULL,
    counter_type VARCHAR(50) NOT NULL,
    last_value BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    PRIMARY KEY (company_id, counter_type),
    CONSTRAINT fk_company_counters_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_company_counters_company_id ON company_counters (company_id);
