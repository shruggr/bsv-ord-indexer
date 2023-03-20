ALTER TABLE inscriptions ADD COLUMN lock BYTEA;

CREATE INDEX idx_inscriptions_lock ON inscriptions(lock);