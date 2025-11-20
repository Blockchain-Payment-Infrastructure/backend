-- Add updated_at and last_password_changed_at to users table
-- Add columns only if they don't already exist to make this migration idempotent
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS last_password_changed_at TIMESTAMP;

-- Optionally populate last_password_changed_at with created_at for existing users
UPDATE users SET last_password_changed_at = created_at WHERE last_password_changed_at IS NULL;
