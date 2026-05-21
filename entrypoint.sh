#!/bin/sh

echo "Running database migrations..."
goose -dir /app/migrations postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" up

echo "Starting application..."
exec /app/app