-- migrations/000003_add_user_id.down.sql
ALTER TABLE url DROP COLUMN IF EXISTS is_deleted;