CREATE TABLE IF NOT EXISTS feeds(
  id uuid NOT NULL PRIMARY KEY,
  title text NOT NULL,
  url text,
  text text,
  type char CHECK (type in ('M', 'J', 'A', 'S')),
  by varchar(16) NOT NULL,
  points smallint NOT NULL,
  n_comments smallint NOT NULL,
  created_at timestamptz NOT NULL
);
