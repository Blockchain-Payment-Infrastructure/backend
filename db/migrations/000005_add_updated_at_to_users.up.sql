ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();

ALTER TABLE users ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMP;

UPDATE users SET password_changed_at = created_at  WHERE password_changed_at IS NULL;

UPDATE users SET updated_at = created_at;
