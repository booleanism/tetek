CREATE TABLE IF NOT EXISTS points(
  id uuid NOT NULL PRIMARY KEY,
  by varchar(16) NOT NULL,
  comments_id uuid NOT NULL,

  CONSTRAINT point_per_user UNIQUE (by, comments_id)
);
