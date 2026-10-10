#!/bin/bash
set -e
echo "=== Social Video Cutter Dev Setup ==="

# Copy .env if not exists
if [ ! -f .env ]; then
  cp .env.example .env
  echo "Created .env from .env.example"
fi

# Build API
echo "--- Building API..."
cd apps/api
go mod tidy
go build ./...
echo "API: OK"
cd ../..

# Build Worker
echo "--- Building Worker..."
cd workers/media-worker
go mod tidy
go build ./...
echo "Worker: OK"
cd ../..

# Install and build Web
echo "--- Building Web..."
cd apps/web
npm install
npm run build
echo "Web: OK"
cd ../..

echo ""
echo "=== All components built successfully! ==="
echo ""
echo "To run with Docker:"
echo "  docker compose up --build"
echo ""
echo "To run locally (requires postgres + redis):"
echo "  cd apps/api && go run ./cmd/server"
echo "  cd workers/media-worker && go run ./cmd/worker"
echo "  cd apps/web && npm run dev"
