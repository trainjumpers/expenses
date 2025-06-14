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
    ('Coffee Shop', 'Morning coffee', 5.50, '2024-01-09', 1, 1);

-- Transaction category mappings for user 1
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

-- Set sequence to continue from the last inserted ID
SELECT setval('test.transaction_id_seq', 10, true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM test.transaction_category_mapping;
DELETE FROM test.transaction;
-- +goose StatementEnd 