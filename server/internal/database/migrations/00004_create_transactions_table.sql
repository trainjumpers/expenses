-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ${DB_SCHEMA}.transaction (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    amount DECIMAL(15, 2) NOT NULL,
    date DATE NOT NULL,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_transaction_created_by FOREIGN KEY (created_by) REFERENCES ${DB_SCHEMA}.user(id)
);

-- Create index for better query performance
CREATE INDEX idx_transaction_created_by ON ${DB_SCHEMA}.transaction(created_by);
CREATE INDEX idx_transaction_date ON ${DB_SCHEMA}.transaction(date);
CREATE INDEX idx_transaction_deleted_at ON ${DB_SCHEMA}.transaction(deleted_at);

CREATE TRIGGER update_transaction_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.transaction
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

-- Create composite unique constraint to prevent duplicate transactions
-- This ensures no duplicate transactions for the same user with identical: date, name, description, and amount
-- Using COALESCE to handle NULL descriptions consistently
CREATE UNIQUE INDEX idx_transaction_unique_composite ON ${DB_SCHEMA}.transaction(
    created_by, 
    date, 
    name, 
    COALESCE(description, ''), 
    amount
) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS ${DB_SCHEMA}.idx_transaction_unique_composite;
DROP INDEX IF EXISTS ${DB_SCHEMA}.idx_transaction_deleted_at;
DROP INDEX IF EXISTS ${DB_SCHEMA}.idx_transaction_date;
DROP INDEX IF EXISTS ${DB_SCHEMA}.idx_transaction_created_by;
DROP TABLE IF EXISTS ${DB_SCHEMA}.transaction;
DROP TRIGGER IF EXISTS update_transaction_modtime ON ${DB_SCHEMA}.transaction;
-- +goose StatementEnd
