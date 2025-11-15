#!/usr/bin/env bash
set -euo pipefail

# 1. Get Postgres container IP
PG_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' postgres_container)

# 2. Get DB credentials from the container env
POSTGRES_USER=$(docker exec postgres_container printenv POSTGRES_USER)
POSTGRES_PASSWORD=$(docker exec postgres_container printenv POSTGRES_PASSWORD)
POSTGRES_DB=$(docker exec postgres_container printenv POSTGRES_DB)

# 3. Build DB URL
DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${PG_IP}:5432/${POSTGRES_DB}?sslmode=disable"

echo "Running migrations with goose against ${DB_URL}"

# 4. Run goose migrations
goose -dir=../db/migrations postgres "${DB_URL}" up