ALTER TABLE IF EXISTS hiddenfeeds ADD CONSTRAINT unique_hidden_per_user UNIQUE (to_uname, feed);
