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

CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.user
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE UNIQUE INDEX unique_active_email
ON ${DB_SCHEMA}.user (email)
WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS ${DB_SCHEMA}.unique_active_email;
DROP TRIGGER IF EXISTS update_user_modtime ON ${DB_SCHEMA}.user;
DROP TABLE IF EXISTS ${DB_SCHEMA}.user;
-- +goose StatementEnd
