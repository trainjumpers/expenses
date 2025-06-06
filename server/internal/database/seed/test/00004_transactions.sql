-- +goose Up
-- +goose StatementBegin
-- Test transactions for user 1
INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (1, 'Integration Transaction', 'Test Description', 100.50, '2023-01-01', 1, NOW(), NOW());

INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (2, 'Transaction without description', NULL, 75.25, '2023-01-02', 1, NOW(), NOW());

INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (3, 'Transaction for Update Test', 'Update Description', 200.00, '2023-01-03', 1, NOW(), NOW());

INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (4, 'Transaction for Delete Test', 'Delete Description', 250.00, '2023-01-04', 1, NOW(), NOW());

-- Test transactions for user 2 (for cross-user access tests)
INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (5, 'User 2 Transaction', 'User 2 Description', 300.00, '2023-01-05', 2, NOW(), NOW());

INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (6, 'User 2 Transaction 2', 'User 2 Description 2', 150.00, '2023-01-06', 2, NOW(), NOW());

-- Additional transactions for listing tests
INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (7, 'List Transaction A', 'List Description A', 100.00, '2023-01-07', 1, NOW(), NOW());

INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (8, 'List Transaction B', 'List Description B', 150.00, '2023-01-08', 1, NOW(), NOW());

INSERT INTO test.transaction (id, name, description, amount, date, created_by, created_at, updated_at) 
VALUES (9, 'List Transaction C', 'List Description C', 200.00, '2023-01-09', 1, NOW(), NOW());

-- Set sequence to continue from the last inserted ID
SELECT setval('test.transaction_id_seq', 9, true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM test.transaction;
-- +goose StatementEnd 