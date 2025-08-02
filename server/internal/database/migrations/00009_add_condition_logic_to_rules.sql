-- +goose Up
-- +goose StatementBegin
ALTER TABLE ${DB_SCHEMA}.rule
ADD COLUMN condition_logic VARCHAR(3) NOT NULL DEFAULT 'AND' CHECK (condition_logic IN ('AND', 'OR'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ${DB_SCHEMA}.rule
DROP COLUMN condition_logic;
-- +goose StatementEnd
