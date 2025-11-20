-- Revert added columns
ALTER TABLE users
    DROP COLUMN IF EXISTS last_password_changed_at,
    DROP COLUMN IF EXISTS updated_at;
