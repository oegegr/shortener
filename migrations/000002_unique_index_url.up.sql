-- migrations/000002_unique_index_url.up.sql
CREATE UNIQUE INDEX idx_unique_url ON url(url);