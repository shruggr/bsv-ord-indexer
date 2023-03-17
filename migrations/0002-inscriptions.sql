CREATE TABLE progress(
    indexer VARCHAR(32) PRIMARY KEY,
    height INTEGER
);

CREATE TABLE inscriptions(
    id BIGINT,
    txid BYTEA,
    vout INTEGER,
    sat BIGINT,
    height INTEGER,
    idx INTEGER,
    origin BYTEA,
    ordinal BIGINT,
    PRIMARY KEY(txid, vout, sat)
);

CREATE INDEX idx_inscriptions_id
ON inscriptions(id);

CREATE INDEX idx_inscriptions_origin
ON inscriptions(origin);

CREATE INDEX idx_inscriptions_ordinal
ON inscriptions(ordinal);