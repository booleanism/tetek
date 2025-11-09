CREATE TABLE IF NOT EXISTS hiddenfeeds(
  id uuid NOT NULL PRIMARY KEY,
  to_uname varchar(16) NOT NULL,
  feed uuid NOT NULL,
  CONSTRAINT fk_feed FOREIGN KEY (feed) REFERENCES feeds (id)
);
