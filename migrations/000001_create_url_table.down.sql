-- migrations/000001_create_url_table.down.sql
DROP INDEX IF EXISTS idx_short_id;
DROP TABLE IF EXISTS url;