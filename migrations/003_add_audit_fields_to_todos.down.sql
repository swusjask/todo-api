-- migrations/003_add_audit_fields_to_todos.down.sql
-- Remove audit fields from todos table

-- Drop indexes first
DROP INDEX IF EXISTS idx_todos_created_by;
DROP INDEX IF EXISTS idx_todos_updated_by;

-- Drop foreign key constraints
ALTER TABLE todos 
DROP CONSTRAINT IF EXISTS fk_todos_created_by,
DROP CONSTRAINT IF EXISTS fk_todos_updated_by;

-- Drop columns
ALTER TABLE todos 
DROP COLUMN IF EXISTS created_by,
DROP COLUMN IF EXISTS updated_by;