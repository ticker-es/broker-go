# broker-go

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