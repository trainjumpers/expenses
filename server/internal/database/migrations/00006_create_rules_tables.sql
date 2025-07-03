-- +goose Up
-- +goose StatementBegin
CREATE TABLE ${DB_SCHEMA}.rule (
   id SERIAL PRIMARY KEY,
   name VARCHAR(100) NOT NULL,
   description TEXT NULL,
   effective_from TIMESTAMP NOT NULL,
   created_by INTEGER NOT NULL,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
   deleted_at TIMESTAMPTZ NULL,
   CONSTRAINT fk_created_by FOREIGN KEY (created_by) REFERENCES ${DB_SCHEMA}.user(id)
);

CREATE TABLE ${DB_SCHEMA}.rule_action (
   id SERIAL PRIMARY KEY,
   rule_id INTEGER NOT NULL,
   action_type VARCHAR(100) NOT NULL,
   action_value VARCHAR(100) NOT NULL,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
   CONSTRAINT fk_rule FOREIGN KEY (rule_id) REFERENCES ${DB_SCHEMA}.rule(id) ON DELETE CASCADE
);

CREATE TABLE ${DB_SCHEMA}.rule_condition (
   id SERIAL PRIMARY KEY,
   rule_id INTEGER NOT NULL,
   condition_type VARCHAR(100) NOT NULL,
   condition_value VARCHAR(100) NOT NULL,
   condition_operator VARCHAR(100) NOT NULL,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
   CONSTRAINT fk_rule FOREIGN KEY (rule_id) REFERENCES ${DB_SCHEMA}.rule(id) ON DELETE CASCADE
);

CREATE TRIGGER update_rules_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.rule
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_rule_action_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.rule_action
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();

CREATE TRIGGER update_rule_condition_modtime
BEFORE UPDATE ON ${DB_SCHEMA}.rule_condition
FOR EACH ROW
EXECUTE FUNCTION update_modified_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_rule_condition_modtime ON ${DB_SCHEMA}.rule_condition;
DROP TRIGGER IF EXISTS update_rule_action_modtime ON ${DB_SCHEMA}.rule_action;
DROP TRIGGER IF EXISTS update_rules_modtime ON ${DB_SCHEMA}.rule;
DROP TABLE IF EXISTS ${DB_SCHEMA}.rule_condition;
DROP TABLE IF EXISTS ${DB_SCHEMA}.rule_action;
DROP TABLE IF EXISTS ${DB_SCHEMA}.rule;
-- +goose StatementEnd 