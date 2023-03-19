CREATE TABLE IF NOT EXISTS blocks(
	id BYTEA PRIMARY KEY,
	height INTEGER,
	indexed TIMESTAMP,
	status INTEGER,
	created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_blocks_height_status
ON blocks(height, status);

INSERT INTO blocks(id, height, status, indexed)
VALUES(
	decode('000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f', 'hex')),
	0,
	CURRENT_TIMESTAMP,
	2
)
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_blocks_status_height
ON blocks(status, height DESC);

CREATE TABLE IF NOT EXISTS txos(
	txid BYTEA, -- 32
	vout INTEGER, -- 4
	satoshis BIGINT, -- 8
	scripthash BYTEA, -- 32
	coinbase INTEGER, -- 4 - block height if transaction is a coinbase transaction
	spend_txid BYTEA, -- 32
	spend_vin INTEGER, -- 4 
	PRIMARY KEY(txid, vout)
);

CREATE INDEX IF NOT EXISTS idx_txos_spend_txid_vin
ON txos(spend_txid, spend_vin)
INCLUDE(txid, vout, satoshis, coinbase);

CREATE INDEX IF NOT EXISTS idx_txos_scripthash
ON txos(scripthash);

CREATE INDEX IF NOT EXISTS idx_txos_scripthash_utxos
ON txos(scripthash)
WHERE spend_txid IS NULL;

CREATE TABLE IF NOT EXISTS blk_txns(
	height INTEGER,
	idx INTEGER,
	txid BYTEA,
	fee BIGINT,
	acc BIGINT, -- accumulation of fees in block when summed by height, block
	PRIMARY KEY(height, idx)
);

CREATE INDEX IF NOT EXISTS idx_blk_txns_height_acc 
ON blk_txns(height, acc DESC)
INCLUDE(txid, fee);