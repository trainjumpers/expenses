-- +goose Up
-- +goose StatementBegin
INSERT INTO test.account (name, bank_type, currency, created_by, created_at, updated_at) VALUES ('Test account 1', 'sbi', 'inr', 1, NOW(), NOW());
INSERT INTO test.account (name, bank_type, currency, created_by, created_at, updated_at) VALUES ('Test account 2', 'axis', 'usd', 1, NOW(), NOW());
INSERT INTO test.account (name, bank_type, currency, created_by, created_at, updated_at) VALUES ('Test account 3', 'axis', 'inr', 2, NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM test.account;
-- +goose StatementEnd
