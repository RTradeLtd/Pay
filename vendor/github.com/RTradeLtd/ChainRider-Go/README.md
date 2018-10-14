# ChainRider-Go

Basic wrapper around the ChainRider API

## Usage

The payload is how we authenticate with chainrider, and consists of the following:
`"{\n  \"token\": \"%s\"\n}"`

The url pattern for chain rider is as follows:
`https://api.chainrider.io/<API_VERSION>/<DIGITAL_CURRENCY>/<BLOCKCHAIN>`

Examples:
* `https://api-dot-chainrider.io/v1/dash/main`
* `https://api-dot-chainrider.io/v1/dash/testnet`

## Supported Calls

At the moment, only DASH is supported

### DASH

* `POST /v1/ratelimit/`
* `GET /status?q=getInfo` 
* `GET /tx/<tx_hash>`
* `GET /txns` (partial support)
* `GET /addr/<address>/balance`
* `GET /block/<block_hash>`
* `GET /status?q=getLastBlockHash`
* `GET /sync`