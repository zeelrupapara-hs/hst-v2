#!/bin/sh
# The schema walks forward before the server takes a request; search_path pins the version
# table to public, since the role shares its name with the hst schema.
set -e

DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSL_MODE}&search_path=public"

until migrate -path /migrations -database "$DB_URL" up; do
  echo "migrations not applied yet, database may still be starting; retrying"
  sleep 3
done

exec /hstserver
