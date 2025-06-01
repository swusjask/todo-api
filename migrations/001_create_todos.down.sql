-- migrations/001_create_todos.down.sql
-- This file reverses the changes made by the .up.sql file
-- It's important for rolling back if something goes wrong

DROP TRIGGER IF EXISTS update_todos_updated_at ON todos;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_todos_created_at;
DROP TABLE IF EXISTS todos;