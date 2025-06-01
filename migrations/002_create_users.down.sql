
-- migrations/002_create_users.down.sql
-- Rollback users table creation

-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables (refresh_tokens first due to foreign key)
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;