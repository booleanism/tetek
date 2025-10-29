CREATE TABLE IF NOT EXISTS users(
  id uuid NOT NULL PRIMARY KEY,
  uname varchar(16) NOT NULL UNIQUE,
  email varchar(64) NOT NULL UNIQUE,
  passwd text NOT NULL,
  role char CHECK (role in ('M', 'N')) NOT NULL,
  created_at timestamptz NOT NULL
);

