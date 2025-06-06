-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ${DB_SCHEMA}.categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(100) NULL,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES ${DB_SCHEMA}.user(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_category_name_created_by
    ON ${DB_SCHEMA}.categories (name, created_by);

CREATE TRIGGER update_categories_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.categories
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_categories_modtime ON ${DB_SCHEMA}.categories;
DROP TABLE IF EXISTS ${DB_SCHEMA}.categories; 
-- +goose StatementEnd
