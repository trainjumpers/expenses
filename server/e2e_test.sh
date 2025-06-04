#!/usr/bin/env bash
set -euo pipefail                             # fail fast, fail on unset vars  

unset DB_SCHEMA
unset DB_SEED_DIR

export DB_SCHEMA=${DB_SCHEMA:-test}
export DB_SEED_DIR=${DB_SEED_DIR:-./internal/database/seed/test}   # ← no trailing space!
export ENV=${ENV:-test}

if [[ "$DB_SCHEMA" != "test" ]]; then
  echo "Refusing to run e2e tests on non-test schema $DB_SCHEMA"
  exit 1
fi                                

cleanup() {
  echo "Cleaning up..."
  echo "y" | just db-downgrade-reset || true
  kill -- -$$
  echo "Done"
}
trap cleanup EXIT                                     

echo "Running migrations into $DB_SCHEMA/$DB_NAME"
just db-upgrade                              
just db-seed

ginkgo -r -p -race ./...
