-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ${DB_SCHEMA}.account (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    balance DECIMAL(10, 2) NOT NULL DEFAULT 0,
    bank_type VARCHAR(100) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES ${DB_SCHEMA}.user(id)
);

CREATE TRIGGER update_account_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.account
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_account_modtime ON ${DB_SCHEMA}.account;
DROP TABLE IF EXISTS ${DB_SCHEMA}.account;
-- +goose StatementEnd
