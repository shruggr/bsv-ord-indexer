# bsv-ord-indexer

## Ordinal Indexing
*Not yet functioning.*

Ordinal Indexing is a means of walking back the blockchain to determin a unique serial number (ordinal) assigned to a single satoshi. Details of ordinals can be found here: [https://docs.ordinals.com/]

Indexing Ordinals can be a heavy process without some specialized indexing of the blockchain. These indexes will be built in the near future

## 1Sat Origin Indexing
The BSV blockchain is unique amount blockchains which support ordinals, in that BSV supports single satoshi outputs. This allows us to take some short-cuts in indexing until a full ordinal indexer can be built efficiently. 

Since ordinals are a unique serial number for each satoshi, an `origin` can be defined as the first outpoint where a satoshi exists alone, in a one satoshi output. Each subsequent spend of that satoshi will be crawled back only to the first ancestor where the origin has already been identified, or until it's ancestor is an output which contains more than one satoshi.

If a satoshi is subsequently packaged up in an output of more than one satoshi, the origin is no longer carried forward. If the satoshi is later spent into another one satoshi output, a new origin will be created. Both of these origins would be the same ordinal, and when the ordinal indexer is complete, both those origins will be identified as both being the same ordinal.

### Environment Variables
- POSTGRES=`<postgres connection string>`
- LISTEN=`<IP>`:`<PORT>` # defaults to 0.0.0.0:8080
- ONESAT=`<GorillaPool "1sat" Subscription ID>`


## Run DB migrations
```
cd migrations
go build
./migrations
```

## Run Web Server
```
cd server
go build
./server
```

## Run Inscriptions Indexer
```
cd 1sat-listener
go build
./1sat-listener
```

## APIs
#### GET `/api/origins/:txid/:vout`
- Returns `origin` as a hex string

#### GET `/api/files/origins/:origin`
- Returns inscribed file

#### POST `/api/inscriptions/:txid`
- Payload: none
- Force indexing of a transaction






