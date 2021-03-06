CREATE TABLE IF NOT EXISTS visits(
   id serial PRIMARY KEY,
   redirect_id INT NOT NULL,
   ip_address text NOT NULL,
   country text NOT NULL,
   region_name text NOT NULL,
   city text NOT NULL,
   isp text NOT NULL,
   created_at TIMESTAMP default CURRENT_TIMESTAMP,
   CONSTRAINT fk_redirect
     FOREIGN KEY(redirect_id)
     REFERENCES redirects(id)
     ON DELETE RESTRICT
);

CREATE INDEX idx_visits_redirect_id ON visits(redirect_id);
