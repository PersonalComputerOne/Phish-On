-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "source" (
    "id" SERIAL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "url" TEXT NOT NULL UNIQUE,
    "last_crawled_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "added_at" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "source";
-- +goose StatementEnd
