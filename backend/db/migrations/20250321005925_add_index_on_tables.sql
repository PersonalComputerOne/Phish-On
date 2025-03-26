-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS domain_url_idx ON "domain" (url);
CREATE INDEX IF NOT EXISTS domain_source_idx ON "domain" (source_id);
-- +goose StatementEnd
