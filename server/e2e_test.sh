#!/usr/bin/env bash
set -euo pipefail                             # fail fast, fail on unset vars  

unset DB_SCHEMA
unset DB_SEED_DIR

export DB_SCHEMA=${DB_SCHEMA:-test}
export DB_SEED_DIR=${DB_SEED_DIR:-./internal/database/seed/test}   # ← no trailing space!

echo "DB_SCHEMA: $DB_SCHEMA"
echo "DB_SEED_DIR: $DB_SEED_DIR"
echo "DB_NAME: $DB_NAME"

if [[ "$DB_SCHEMA" != "test" ]]; then
  echo "❌ Refusing to run e2e tests on non-test schema $DB_SCHEMA"
  exit 1
fi                                

cleanup() {
  echo "Cleaning up..."
  echo "y" | just db-downgrade-reset || true
  kill -- -$$
  echo "Done ✔"
}
trap cleanup EXIT                                             

#####--- Migrate + seed ─────────────
echo "⏫ Running migrations into $DB_SCHEMA/$DB_NAME"
just db-upgrade                                                
just db-seed

#####--- Start the API server ───────
go run cmd/neurospend/main.go &  SERVER_PID=$!
echo "API started (pid $SERVER_PID); waiting for healthy status…"

for _ in {1..20}; do                                          
  if curl -fs "http://localhost:${SERVER_PORT}/health" >/dev/null; then
    echo "✅ Server healthy"
    break
  fi
  sleep 1
done

#####--- Run the tests ─────────────
ginkgo -r ./                                                   
