# StreamingFast GraphQL API Services for Blockchains
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/streamingfast/dgraphql)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This service is the API for querying blockchain data using GraphQL.
It is part of **[StreamingFast](https://github.com/streamingfast/streamingfast)**.


## Installation

See the different protocol-specific `StreamingFast` binaries at https://github.com/streamingfast/streamingfast#protocols

Current `dgraphql` implementations:

* [**StreamingFast on EOSIO**](https://github.com/streamingfast/sf-eosio)
* [**StreamingFast on Ethereum**](https://github.com/streamingfast/sf-ethereum)

## Usage

## Connecting from command-line clients

* Using https://github.com/Gavinpub/ws (which supports subprotocol negotiation)

    ws -s graphql-ws ws://127.0.0.1:8080/graphql

    {"id":"3","type":"connection_init","payload":{"Authorization":"Bearer ......"},"variables":null}}

    {"id":"3","type":"start","payload":{"query":"subscription{searchTransactionsForward(query: \"status:executed\", lowBlockNum:0, highBlockNum:0){ trace{ block { num } id }}}","variables":null}}

    {"id":"3","type":"start","payload":{"query":"subscription{searchTransactionsForward(query: \"status:executed\", lowBlockNum:0, highBlockNum:0){ trace{ block { num } id }}}","variables":null}}

* Using curl:

    curl http://localhost:8080/graphql -XPOST -d '{"query": "{ searchTransactionsForward(limit: 10, query: \"status:executed\") { cursor } }"}' -H "Authorization: Bearer $SF_API_TOKEN" | jq .

    curl http://localhost:8080/graphql -XPOST -d '{"query": "{ blockIDByTime(time: \"2019-01-01T00:00:00Z\") { time num } }"}' -H "Authorization: Bearer $SF_API_TOKEN" | jq .

## Contributing

**Issues and PR in this repo related strictly to the core dgraphql API engine**

Report any protocol-specific issues in their
[respective repositories](https://github.com/streamingfast/streamingfast#protocols)

**Please first refer to the general
[StreamingFast contribution guide](https://github.com/streamingfast/streamingfast/blob/master/CONTRIBUTING.md)**,
if you wish to contribute to this code base.

This codebase uses unit tests extensively, please write and run tests.


## License

[Apache 2.0](LICENSE)
