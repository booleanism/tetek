CREATE TABLE IF NOT EXISTS comments(
  id uuid NOT NULL PRIMARY KEY,
  parent uuid NOT NULL,
  text TEXT NOT NULL,
  by varchar(16) NOT NULL,
  created_at timestamptz NOT NULL
);
