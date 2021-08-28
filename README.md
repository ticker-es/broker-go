# Ticker Broker

Implements an EventStream broker with multiple backends for storing Events and Subscriptions. 

## Usage

Full help available with

`ticker-broker help`

Start a broker with `postgres` backend for storing Events but using `redis` for storing subscriptions:

`ticker-broker server --event-store postgres --evt-postgres-url=<PG_URL> --sequence-store redis 
--seq-redis-url=<REDIS_URL>`

## Development

### Generate Certificates

Create CA for signing server and client certificates:

`certstrap --depot-path . init --passphrase '' --cn 'ca'`

Create and sign certificate for server

`certstrap --depot-path . request-cert --passphrase '' --domain 'localhost' --cn 'broker'`
`certstrap --depot-path . sign --passphrase '' --CA 'ca' broker`

Create and sign client certificate

`certstrap --depot-path . request-cert --passphrase '' --ip '127.0.0.1' --cn 'client-1'`
`certstrap --depot-path . sign --passphrase '' --CA 'ca' client-1`