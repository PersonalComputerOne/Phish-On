-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "suspicious_links" (
    "id" SERIAL PRIMARY KEY,
    "url" TEXT NOT NULL UNIQUE,
    "added_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd
