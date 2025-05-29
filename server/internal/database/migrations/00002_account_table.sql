-- +goose Up
-- SQL migration for creating the accounts table

CREATE TABLE IF NOT EXISTS ${DB_SCHEMA}.account (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    balance FLOAT NOT NULL DEFAULT 0,
    bank_type VARCHAR(100) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES ${DB_SCHEMA}.user(id)
);

-- +goose Down
DROP TABLE IF EXISTS ${DB_SCHEMA}.account; 