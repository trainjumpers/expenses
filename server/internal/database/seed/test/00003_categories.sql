-- +goose Up
-- +goose StatementBegin
INSERT INTO test.categories (name, icon, created_by, created_at, updated_at) VALUES ('Food', 'burger-icon', 1, NOW(), NOW());
INSERT INTO test.categories (name, icon, created_by, created_at, updated_at) VALUES ('Transportation', 'car-icon', 1, NOW(), NOW());
INSERT INTO test.categories (name, icon, created_by, created_at, updated_at) VALUES ('Entertainment', 'entertainment-icon', 1, NOW(), NOW());
INSERT INTO test.categories (name, icon, created_by, created_at, updated_at) VALUES ('Shopping', 'shopping-icon', 1, NOW(), NOW());
INSERT INTO test.categories (name, icon, created_by, created_at, updated_at) VALUES ('Health', 'health-icon', 1, NOW(), NOW());
INSERT INTO test.categories (name, icon, created_by, created_at, updated_at) VALUES ('Other', 'other-icon', 2, NOW(), NOW());
INSERT INTO test.categories (name, icon, created_by, created_at, updated_at) VALUES ('Salary', 'salary-icon', 2, NOW(), NOW())
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd 