-- +goose Up
-- +goose StatementBegin
CREATE TABLE ${DB_SCHEMA}.rule_transaction_mapping (
   id SERIAL PRIMARY KEY,
   rule_id INTEGER NOT NULL,
   transaction_id INTEGER NOT NULL,
   applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
   CONSTRAINT fk_rule_mapping FOREIGN KEY (rule_id) REFERENCES ${DB_SCHEMA}.rule(id) ON DELETE CASCADE,
   CONSTRAINT fk_transaction_mapping FOREIGN KEY (transaction_id) REFERENCES ${DB_SCHEMA}.transaction(id) ON DELETE CASCADE,
   CONSTRAINT unique_rule_transaction UNIQUE (rule_id, transaction_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ${DB_SCHEMA}.rule_transaction_mapping;
-- +goose StatementEnd