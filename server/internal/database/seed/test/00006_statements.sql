-- +goose Up
-- +goose StatementBegin
-- Seed statements for integration tests (account_id = 1, created_by = 1)
INSERT INTO test.statement (account_id, created_by, original_filename, file_type, status, message, created_at, updated_at)
VALUES
    (1, 1, 'salary_jan.csv', 'csv', 'done', 'Seed Salary statement', NOW(), NOW()),
    (1, 1, 'groceries_feb.csv', 'csv', 'done', 'Seed Groceries statement', NOW(), NOW()),
    (1, 1, 'utilities_mar.csv', 'csv', 'done', 'Seed Utilities statement', NOW(), NOW()),
    (3, 2, 'user2_statement.csv', 'csv', 'done', 'Seed User2 statement', NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM test.statement;
-- +goose StatementEnd
