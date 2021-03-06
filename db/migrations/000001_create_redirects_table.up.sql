CREATE TABLE IF NOT EXISTS redirects(
   id serial PRIMARY KEY,
   link text NOT NULL,
   url text NOT NULL,
   access_code text NOT NULL default md5(random()::text),
   created_at TIMESTAMP default CURRENT_TIMESTAMP,
   deleted_at TIMESTAMP
);

CREATE INDEX idx_redirects_link ON redirects(link);
CREATE INDEX idx_redirects_access_code ON redirects(access_code);
