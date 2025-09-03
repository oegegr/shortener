-- migrations/000003_add_user_id.up.sql
ALTER TABLE url ADD COLUMN is_deleted BOOLEAN;
