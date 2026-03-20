-- +goose Up
SELECT 'up SQL query';

-- Create a grinders table
CREATE TABLE IF NOT EXISTS grinders (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    version int NOT NULL DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

-- Add grinder_id to recipes
ALTER TABLE recipes
ADD COLUMN IF NOT EXISTS grinder_id bigint REFERENCES grinders(id) ON DELETE SET NULL;

-- +goose Down
SELECT 'down SQL query';

-- Remove column
ALTER TABLE recipes
DROP COLUMN IF EXISTS grinder_id;

-- Drop grinders table
DROP TABLE IF EXISTS grinders;
