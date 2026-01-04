-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ${DB_SCHEMA}.investment_account_value (
    account_id INTEGER PRIMARY KEY,
    current_value DECIMAL(14, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_investment_account_value_account_id FOREIGN KEY (account_id) REFERENCES ${DB_SCHEMA}.account(id) ON DELETE CASCADE
);

CREATE TRIGGER update_investment_account_value_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.investment_account_value
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_investment_account_value_modtime ON ${DB_SCHEMA}.investment_account_value;
DROP TABLE IF EXISTS ${DB_SCHEMA}.investment_account_value;
-- +goose StatementEnd
