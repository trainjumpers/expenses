-- +goose Up
-- +goose StatementBegin
CREATE TABLE ${DB_SCHEMA}.job (
    id SERIAL PRIMARY KEY,
    job_type VARCHAR(50) NOT NULL, -- 'rule_execution', 'statement_processing', etc.
    reference_id INTEGER, -- can reference rule_id, statement_id, etc. depending on job_type
    created_by INTEGER NOT NULL REFERENCES ${DB_SCHEMA}."user"(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    message TEXT,
    metadata JSONB, -- flexible field for job-specific data
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_job_type_status ON ${DB_SCHEMA}.job(job_type, status);
CREATE INDEX idx_job_created_by ON ${DB_SCHEMA}.job(created_by);
CREATE INDEX idx_job_reference_id ON ${DB_SCHEMA}.job(reference_id);

CREATE TRIGGER update_job_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.job
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_job_modtime ON ${DB_SCHEMA}.job;
DROP INDEX IF EXISTS ${DB_SCHEMA}.idx_job_reference_id;
DROP INDEX IF EXISTS ${DB_SCHEMA}.idx_job_created_by;
DROP INDEX IF EXISTS ${DB_SCHEMA}.idx_job_type_status;
DROP TABLE IF EXISTS ${DB_SCHEMA}.job;
-- +goose StatementEnd