#!/usr/bin/env bash
set -euo pipefail

export DB_SCHEMA="test"
export DB_SEED_DIR=${DB_SEED_DIR:-./internal/database/seed/test}
export ENV="test"

if [[ "$DB_SCHEMA" != "test" ]]; then
  echo "Refusing to run e2e tests on non-test schema $DB_SCHEMA"
  exit 1
fi

cleanup() {
  echo "Cleaning up..."
  echo "y" | just db-downgrade-reset || true
}
trap cleanup EXIT

echo "Running migrations into $DB_SCHEMA/$DB_NAME"
just db-upgrade
just db-seed

FOCUS_STRING=""
if [[ $# -gt 0 ]]; then
  FOCUS_STRING="$1"
fi
echo "Running e2e tests with focus: $FOCUS_STRING"
ginkgo -r -race -cover -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... --focus "$FOCUS_STRING" ./...
