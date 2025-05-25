-- +goose Up
-- +goose StatementBegin
INSERT INTO test.user (name, email, password, created_at, updated_at) VALUES ('John Doe', 'john.doe@example.com', 'password', NOW(), NOW());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose StatementEnd
