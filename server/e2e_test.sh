#!/usr/bin/env bash
set -euo pipefail                             # fail fast, fail on unset vars  

unset DB_SCHEMA
unset DB_SEED_DIR

export DB_SCHEMA=${DB_SCHEMA:-test}
export DB_SEED_DIR=${DB_SEED_DIR:-./internal/database/seed/test}   # ← no trailing space!

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

SERVER_HEALTHY=false
for _ in {1..20}; do                                          
  if curl -fs "http://localhost:${SERVER_PORT}/health" >/dev/null; then
    echo "✅ Server healthy"
    SERVER_HEALTHY=true
    break
  fi
  sleep 1
done

if [ "$SERVER_HEALTHY" = false ]; then
  echo "❌ Server failed to become healthy within timeout period"
  exit 1
fi

#####--- Run the tests ─────────────
ginkgo -r ./                                                   
