set unstable

# Database connection string with environment variables
DB_CONNECTION := "user=${DB_USER} dbname=${DB_NAME} password=${DB_PASSWORD} host=${DB_HOST} port=${DB_PORT} sslmode=${DB_SSL_MODE:-verify-full}"
# Database driver for migrations
GOOSE_DRIVER := "postgres"
# Table to track migration versions
GOOSE_TABLE := "${DB_SCHEMA}.goose_db_version"
# Directory containing migration files
GOOSE_MIGRATION_DIR := "./internal/database/migrations"
# Goose Seed directory. Seed data is used mostly for testing.
GOOSE_SEED_DIR := "${DB_SEED_DIR:-./internal/database/seed/prod}"

# Default version to downgrade to when resetting database
default_downgrade := "0"

# Executes the goose command
[private]
[script("bash")]
[working-directory: 'server']
goose command migration_dir:
  GOOSE_DBSTRING="{{DB_CONNECTION}}" GOOSE_DRIVER={{GOOSE_DRIVER}} GOOSE_MIGRATION_DIR={{migration_dir}} GOOSE_TABLE={{GOOSE_TABLE}} goose {{command}}

# Migration commands
# Private helper function to execute database migrations
# Replaces ${DB_SCHEMA} placeholder in migration files with actual schema value
# Executes the goose command and then restores original migration files
[private]
[script("bash")]
[working-directory: 'server']
goose_migrate command:
  find {{GOOSE_MIGRATION_DIR}} -type f -exec sed -i.bak "s|\${DB_SCHEMA}|${DB_SCHEMA}|g" {} +
  just goose "{{command}}" "{{GOOSE_MIGRATION_DIR}}"
  find {{GOOSE_MIGRATION_DIR}} -type f -name "*.bak" | while read -r bak_file; do
  original_file="${bak_file%.bak}"
  mv "$bak_file" "$original_file"
  done

# Creates a new migration file with the specified name
[working-directory: 'server']
db-create file:
  just goose create {{file}} sql

# Applies all pending migrations to upgrade the database schema
db-upgrade:
  just goose_migrate "up"

# Applies seed data to the database
db-seed:
  just goose "up --no-versioning" {{GOOSE_SEED_DIR}}

# Reverts the most recent migration
db-downgrade:
  just goose_migrate "down"

# Downgrades the database to a specific version (defaults to 0)
[script("bash")]
db-downgrade-reset reset=default_downgrade:
  echo "Are you sure you want to reset the database? (y/n)"; read answer;
  if [ "$answer" = "y" ]; then
  just goose_migrate "down-to {{reset}}"
  else
  echo "Database reset cancelled."
  fi

# Installs required Go tools for development (air for hot-reload, goose for migrations)
[working-directory: 'server']
@install: 
  go install github.com/air-verse/air@latest
  go install github.com/pressly/goose/v3/cmd/goose@latest
  go install github.com/onsi/ginkgo/v2/ginkgo@latest
  go mod tidy

# Starts the server with hot-reload using air
[working-directory: 'server']
@server:
  air

@test:
  cd server && bash e2e_test.sh

@test-coverage:
  cd server && COVERAGE=true bash e2e_test.sh