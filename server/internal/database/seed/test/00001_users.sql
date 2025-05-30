-- +goose Up
-- +goose StatementBegin
INSERT INTO test.user (name, email, password, created_at, updated_at) VALUES ('Test user 1', 'test1@example.com', '$2a$14$N13BsEF2NyxT3qUlFIzmLujxCJcjjdf40PS2Gtcyl.ToGBazk0YVe', NOW(), NOW());
INSERT INTO test.user (name, email, password, created_at, updated_at) VALUES ('Test user 2', 'test2@example.com', '$2a$14$N13BsEF2NyxT3qUlFIzmLujxCJcjjdf40PS2Gtcyl.ToGBazk0YVe', NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
