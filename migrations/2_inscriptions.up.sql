CREATE TABLE progress(
    indexer VARCHAR(32) PRIMARY KEY,
    height INTEGER
);

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