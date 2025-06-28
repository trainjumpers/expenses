-- +goose Up
-- +goose StatementBegin
-- Seed rules for integration tests
INSERT INTO test.rule (id, name, description, effective_from, created_by, created_at, updated_at) VALUES
  (1, 'Amount Rule', 'Matches transaction by amount', '2023-01-01', 1, NOW(), NOW()),
  (2, 'Name Rule', 'Matches transaction by name', '2023-01-01', 1, NOW(), NOW());

-- Actions for rules
INSERT INTO test.rule_action (id, rule_id, action_type, action_value, created_at, updated_at) VALUES
  (1, 1, 'name', 'Updated by Amount Rule', NOW(), NOW()),
  (2, 2, 'description', 'Updated by Name Rule', NOW(), NOW());

-- Conditions for rules
INSERT INTO test.rule_condition (id, rule_id, condition_type, condition_value, condition_operator, created_at, updated_at) VALUES
  (1, 1, 'amount', '100.50', 'equals', NOW(), NOW()),
  (2, 2, 'name', 'Integration Transaction', 'equals', NOW(), NOW());

-- Set sequence to continue from the last inserted Id
SELECT setval('test.rule_id_seq', 2, true);
SELECT setval('test.rule_action_id_seq', 2, true);
SELECT setval('test.rule_condition_id_seq', 2, true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM test.rule_condition;
DELETE FROM test.rule_action;
DELETE FROM test.rule;
-- +goose StatementEnd
