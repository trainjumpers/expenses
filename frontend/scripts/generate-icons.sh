#!/bin/bash

# Always clone lucide into the frontend/lucide directory
LUCIDE_DIR="$(dirname "$0")/../lucide"
PACKAGE_JSON="$(dirname "$0")/../package.json"

# Extract lucide-react version from package.json (remove ^ or ~ if present)
LUCIDE_TAG=$(grep '"lucide-react"' "$PACKAGE_JSON" | sed -E 's/.*: *"[\^~]?([^"]+)".*/\1/')

rm -rf "$LUCIDE_DIR"
git clone https://github.com/lucide-icons/lucide.git "$LUCIDE_DIR"

if [ -n "$LUCIDE_TAG" ]; then
  cd "$LUCIDE_DIR" || exit 1
  git checkout "$LUCIDE_TAG"
  cd - > /dev/null
fi

npx ts-node --project tsconfig.scripts.json scripts/generate-icons-data.ts
npx prettier --write components/ui/icons-data.ts
rm -rf "$LUCIDE_DIR"
