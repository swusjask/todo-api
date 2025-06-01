-- migrations/003_add_audit_fields_to_todos.up.sql
-- Add audit fields to todos table

-- Add audit columns to todos table
ALTER TABLE todos 
ADD COLUMN created_by INTEGER,
ADD COLUMN updated_by INTEGER;

-- Add foreign key constraints
ALTER TABLE todos
ADD CONSTRAINT fk_todos_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
ADD CONSTRAINT fk_todos_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL;

-- Create indexes for foreign keys for better join performance
CREATE INDEX idx_todos_created_by ON todos(created_by);
CREATE INDEX idx_todos_updated_by ON todos(updated_by);

-- Optional: Update existing todos to have a default user
-- This requires that you have at least one user in the users table
-- UPDATE todos SET created_by = 1, updated_by = 1 WHERE created_by IS NULL;

-- Note: We don't need to add created_at and updated_at as they already exist in the todos table