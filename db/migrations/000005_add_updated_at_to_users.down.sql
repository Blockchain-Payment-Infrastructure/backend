ALTER TABLE users
    DROP COLUMN IF EXISTS password_changed_at,
    DROP COLUMN IF EXISTS updated_at;
