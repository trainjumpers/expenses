-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ${DB_SCHEMA}.session (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES ${DB_SCHEMA}."user"(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT unique_session_token_hash UNIQUE (token_hash)
);

CREATE INDEX idx_session_user_id ON ${DB_SCHEMA}.session (user_id);
CREATE INDEX idx_session_expires_at ON ${DB_SCHEMA}.session (expires_at);

CREATE TRIGGER update_session_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.session
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_session_modtime ON ${DB_SCHEMA}.session;
DROP TABLE IF EXISTS ${DB_SCHEMA}.session;
-- +goose StatementEnd
