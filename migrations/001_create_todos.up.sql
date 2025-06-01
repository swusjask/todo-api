-- migrations/001_create_todos.up.sql
-- This file creates the todos table
-- The .up.sql extension indicates this migration moves the schema "up" or forward

CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,                           -- SERIAL auto-increments in PostgreSQL
    title VARCHAR(200) NOT NULL,                     -- Required field with max length
    description TEXT DEFAULT '',                     -- Optional, defaults to empty string
    completed BOOLEAN DEFAULT FALSE NOT NULL,        -- Track completion status
    completed_at TIMESTAMP,                          -- NULL when not completed
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,     -- Automatically set on insert
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL      -- We'll update this on every change
);

-- Create an index on created_at for efficient sorting
-- This speeds up our List query that orders by created_at DESC
CREATE INDEX idx_todos_created_at ON todos(created_at DESC);

-- Create a trigger to automatically update updated_at
-- This ensures updated_at is always current without remembering to set it in every UPDATE
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_todos_updated_at 
    BEFORE UPDATE ON todos 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();