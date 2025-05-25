-- +goose Up
-- +goose StatementBegin
CREATE TABLE ${DB_SCHEMA}.user (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX unique_active_email
ON ${DB_SCHEMA}.user (email)
WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS ${DB_SCHEMA}.unique_active_email;
DROP TABLE IF EXISTS ${DB_SCHEMA}.user;
-- +goose StatementEnd
