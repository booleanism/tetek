ALTER TABLE IF EXISTS feeds ADD COLUMN IF NOT EXISTS deleted_at timestamptz;
