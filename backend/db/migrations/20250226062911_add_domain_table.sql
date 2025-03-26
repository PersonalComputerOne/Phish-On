-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "domain" (
    "id" SERIAL PRIMARY KEY,
    "url" TEXT NOT NULL UNIQUE,
    "is_phishing" BOOLEAN DEFAULT FALSE,
    "added_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "source_id" INTEGER NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "domain";
-- +goose StatementEnd
