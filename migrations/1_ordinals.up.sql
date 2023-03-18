CREATE TABLE IF NOT EXISTS ordinals(
	outpoint BYTEA PRIMARY KEY,
	origin BYTEA,
	ordinal BIGINT
);
ALTER TABLE ordinals RENAME TO origins;

CREATE TABLE IF NOT EXISTS ordinals(
	outpoint BYTEA,
	outsat BIGINT,
	origin BYTEA,
	ordinal BIGINT,
	PRIMARY KEY(outpoint, outsat)
);

INSERT INTO ordinals
SELECT outpoint, 0, origin
FROM origins;

DROP TABLE origins;