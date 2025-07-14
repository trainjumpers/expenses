-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ${DB_SCHEMA}.statement (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES ${DB_SCHEMA}.account(id),
    created_by INTEGER NOT NULL REFERENCES ${DB_SCHEMA}."user"(id),
    original_filename VARCHAR(255) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL, -- 'pending', 'processing', 'done', 'error'
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS ${DB_SCHEMA}.statement_transaction_mapping (
    id SERIAL PRIMARY KEY,
    statement_id INTEGER NOT NULL REFERENCES ${DB_SCHEMA}.statement(id) ON DELETE CASCADE,
    transaction_id INTEGER NOT NULL REFERENCES ${DB_SCHEMA}.transaction(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT unique_statement_transaction_mapping UNIQUE (statement_id, transaction_id)
);

CREATE TRIGGER update_statement_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.statement
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_statement_modtime ON ${DB_SCHEMA}.statement;
DROP TABLE IF EXISTS ${DB_SCHEMA}.statement_transaction_mapping;
DROP TABLE IF EXISTS ${DB_SCHEMA}.statement;
-- +goose StatementEnd
