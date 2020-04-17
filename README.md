# dfuse GraphQL API Services for Blockchains
[![reference](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://pkg.go.dev/github.com/dfuse-io/dgraphql)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

This service is the API for querying blockchain data using GraphQL.
It is part of **[dfuse](https://github.com/dfuse-io/dfuse)**.


## Installation

See the different protocol-specific `dfuse` binaries at https://github.com/dfuse-io/dfuse#protocols

Current `dgraphql` implementations:

* [**dfuse for EOSIO**](https://github.com/dfuse-io/dfuse-eosio)
* **dfuse for Ethereum**, soon to be open sourced


## Usage

## Connecting from command-line clients

* Using https://github.com/Gavinpub/ws (which supports subprotocol negotiation)

    ws -s graphql-ws ws://127.0.0.1:8080/graphql

    {"id":"3","type":"connection_init","payload":{"Authorization":"Bearer ......"},"variables":null}}

    {"id":"3","type":"start","payload":{"query":"subscription{searchTransactionsForward(query: \"status:executed\", lowBlockNum:0, highBlockNum:0){ trace{ block { num } id }}}","variables":null}}

    {"id":"3","type":"start","payload":{"query":"subscription{searchTransactionsForward(query: \"status:executed\", lowBlockNum:0, highBlockNum:0){ trace{ block { num } id }}}","variables":null}}

* Using curl:

    curl http://localhost:8080/graphql -XPOST -d '{"query": "{ searchTransactionsForward(limit: 10, query: \"status:executed\") { cursor } }"}' -H "Authorization: Bearer $DFUSE" | jq .

    curl http://localhost:8080/graphql -XPOST -d '{"query": "{ blockIDByTime(time: \"2019-01-01T00:00:00Z\") { time num } }"}' -H "Authorization: Bearer $DFUSE" | jq .

## Contributing

**Issues and PR in this repo related strictly to the core dgraphql API engine**

Report any protocol-specific issues in their
[respective repositories](https://github.com/dfuse-io/dfuse#protocols)

**Please first refer to the general
[dfuse contribution guide](https://github.com/dfuse-io/dfuse#contributing)**,
if you wish to contribute to this code base.

This codebase uses unit tests extensively, please write and run tests.


## License

[Apache 2.0](LICENSE)
