-- +goose Up
-- +goose StatementBegin
CREATE TABLE ${DB_SCHEMA}.transaction_category_mapping (
   id SERIAL PRIMARY KEY,
   category_id INTEGER NOT NULL,
   transaction_id INTEGER NOT NULL,
   CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES ${DB_SCHEMA}.categories(id),
   CONSTRAINT fk_transaction FOREIGN KEY (transaction_id) REFERENCES ${DB_SCHEMA}.transaction(id),
   CONSTRAINT unique_category_transaction_mapping UNIQUE (category_id, transaction_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ${DB_SCHEMA}.transaction_category_mapping;
-- +goose StatementEnd
