CREATE TABLE IF NOT EXISTS ordinals(
	txid BYTEA,
	vout INTEGER,
	sat BIGINT,
	origin BYTEA,
	ordinal BIGINT,
	PRIMARY KEY(txid, vout, outsat)
);

CREATE TABLE progress(
    indexer VARCHAR(32) PRIMARY KEY,
    height INTEGER
);

CREATE TABLE txos(
    txid BYTEA,
    vout INTEGER,
	satoshis BIGINT,
	lockhash BYTEA,
	spend BYTEA,
	vin INTEGER,
	origin BYTEA,
	PRIMARY KEY(txid, vout)
)
CREATE INDEX idx_txos_lockhash ON txos(lockhash, spend_txid);


CREATE TABLE inscriptions(
    txid BYTEA,
    vout INTEGER,
    filehash BYTEA,
    filesize INTEGER,
    filetype VARCHAR(256),
    id BIGINT,
    origin BYTEA,
    ordinal BIGINT,
    height INTEGER,
    idx INTEGER,
    PRIMARY KEY(txid, vout)
);

CREATE INDEX idx_inscriptions_id
ON inscriptions(id);

CREATE INDEX idx_inscriptions_origin
ON inscriptions(origin);

CREATE INDEX idx_inscriptions_ordinal
ON inscriptions(ordinal);

CREATE INDEX idx_inscriptions_filehash
ON inscriptions(filehash);