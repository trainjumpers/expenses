-- +goose Up
-- +goose StatementBegin
-- Transactions for user 1 (test1@example.com)
INSERT INTO test.transaction (name, description, amount, date, created_by, account_id)
VALUES
    ('Integration Transaction', 'Test Description', 100.00, '2023-01-01', 1, 1),
    ('Groceries Shopping', 'Monthly groceries', 150.50, '2023-01-02', 1, 1),
    ('Utility Bill', 'Electricity bill', 75.25, '2023-01-03', 1, 2),
    ('Restaurant', 'Dinner with friends', 45.75, '2023-01-03', 1, 1),
    ('Movie Tickets', 'Weekend movie', 30.00, '2023-01-04', 1, 2),
    ('Fuel', 'Car refuel', 60.00, '2023-01-05', 1, 1),
    ('Online Shopping', 'Amazon purchase', 200.00, '2023-01-06', 1, 2),
    ('Medical Expense', 'Doctor visit', 100.00, '2023-01-07', 1, 1),
    ('Internet Bill', 'Monthly internet', 50.00, '2023-01-08', 1, 2),
    ('Coffee Shop', 'Morning coffee', 5.50, '2024-01-09', 1, 1),
    ('Cash Withdrawal', 'ATM withdrawal', 100.00, '2024-01-10', 1, 1);

-- Transactions for user 2 (test2@example.com)
INSERT INTO test.transaction (name, description, amount, date, created_by, account_id)
VALUES
    ('User2 Transaction 1', 'Test transaction for user 2', 50.00, '2023-02-01', 2, 3),
    ('User2 Transaction 2', 'Another test transaction', 75.00, '2023-02-02', 2, 3),
    ('User2 Transaction 3', 'Third test transaction', 25.00, '2023-02-03', 2, 3),
    ('User2 Transaction 4', 'Fourth test transaction', 125.00, '2023-02-04', 2, 3),
    ('User2 Transaction 5', 'Fifth test transaction', 90.00, '2023-02-05', 2, 3),
    ('User2 Transaction 6', 'Sixth test transaction', 110.00, '2023-02-06', 2, 3);

-- Transaction category mappings for user 1 (note: transaction 11 has no categories - uncategorized)
INSERT INTO test.transaction_category_mapping (transaction_id, category_id)
VALUES
    (1, 1),  -- Integration Transaction -> Food
    (2, 4),  -- Groceries -> Shopping
    (3, 2),  -- Utility Bill -> Transportation (closest available)
    (4, 1),  -- Restaurant -> Food
    (5, 3),  -- Movie Tickets -> Entertainment
    (6, 2),  -- Fuel -> Transportation
    (7, 4),  -- Online Shopping -> Shopping
    (8, 5),  -- Medical Expense -> Health
    (9, 2),  -- Internet Bill -> Transportation (closest available)
    (10, 1); -- Coffee Shop -> Food
    -- Transaction 11 (Cash Withdrawal) intentionally has no category mapping - uncategorized

-- Transaction category mappings for user 2
INSERT INTO test.transaction_category_mapping (transaction_id, category_id)
VALUES
    (12, 6),  -- User2 Transaction 1 -> Other
    (13, 7),  -- User2 Transaction 2 -> Salary
    (14, 6),  -- User2 Transaction 3 -> Other
    (15, 7),  -- User2 Transaction 4 -> Salary
    (16, 6),  -- User2 Transaction 5 -> Other
    (17, 7);  -- User2 Transaction 6 -> Salary

-- Set sequence to continue from the last inserted Id
SELECT setval('test.transaction_id_seq', 17, true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM test.transaction_category_mapping;
DELETE FROM test.transaction;
-- +goose StatementEnd
