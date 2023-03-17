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

CREATE TABLE IF NOT EXISTS txos(
	txid BYTEA,
	vout INTEGER,
	satoshis BIGINT,
	coinbase INTEGER,
	spend BYTEA,
	vin INTEGER,
	PRIMARY KEY(txid, vout)
);

CREATE INDEX IF NOT EXISTS idx_txos_spend_vin
ON txos(spend, vin)
INCLUDE(satoshis, coinbase, txid, vout);

CREATE TABLE IF NOT EXISTS blk_txns(
	height INTEGER,
	idx INTEGER,
	txid BYTEA,
	fee BIGINT,
	acc BIGINT,
	PRIMARY KEY(height, idx)
);

CREATE INDEX IF NOT EXISTS idx_blk_txns_height_acc 
ON blk_txns(height, acc DESC)
INCLUDE(txid, fee);