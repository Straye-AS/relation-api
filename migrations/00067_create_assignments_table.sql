-- +goose Up
-- +goose StatementBegin

-- Assignments table for syncing work orders from ERP datawarehouse
-- Read-only table populated by sync from dbo.cw_<company>_assignments
CREATE TABLE assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Datawarehouse references
    dw_assignment_id BIGINT NOT NULL,           -- AssignmentId from DW
    dw_project_id BIGINT NOT NULL,              -- ProjectId from DW (internal reference)
    
    -- Link to local offer (matched via external_reference = project Code)
    offer_id UUID REFERENCES offers(id) ON DELETE SET NULL,
    company_id VARCHAR(50) NOT NULL REFERENCES companies(id),
    
    -- Core assignment fields
    assignment_number VARCHAR(50) NOT NULL,     -- e.g., "2406200"
    description TEXT,
    
    -- Financial field (just FixedPriceAmount for now)
    fixed_price_amount DECIMAL(15,2) DEFAULT 0,
    
    -- Status tracking (enum IDs from DW)
    status_id INT,                              -- Enum_AssignmentStatusId
    progress_id INT,                            -- Enum_AssignmentProgressId
    
    -- Extensibility: store full DW row as JSONB for future use
    dw_raw_data JSONB,
    
    -- Sync metadata
    dw_synced_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Standard timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: one DW assignment per company
    CONSTRAINT uq_assignments_company_dw_id UNIQUE (company_id, dw_assignment_id)
);

-- Indexes for common queries
CREATE INDEX idx_assignments_offer_id ON assignments(offer_id);
CREATE INDEX idx_assignments_company_id ON assignments(company_id);
CREATE INDEX idx_assignments_assignment_number ON assignments(assignment_number);
CREATE INDEX idx_assignments_dw_project_id ON assignments(dw_project_id);

-- Comment on table
COMMENT ON TABLE assignments IS 'ERP assignments (work orders) synced from datawarehouse. Read-only, populated by sync.';
COMMENT ON COLUMN assignments.dw_raw_data IS 'Full DW row stored as JSONB for future extensibility without migrations';
COMMENT ON COLUMN assignments.fixed_price_amount IS 'FixedPriceAmount from DW - primary financial field for aggregation';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS assignments;
-- +goose StatementEnd

