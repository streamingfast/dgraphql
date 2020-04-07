# dfuse GraphQL API services for different chains

## Prerequistes

You need to have `go-bindata` installed to work on this repository. Follow instructions
at https://github.com/kevinburke/go-bindata to install it. It is used when generating
GraphQL schema to pack them automatically into binary format.

## Maintaining
## Align with `nodeos`'s rework

* `parent_action_ordinal` is uint, and 0 is the implicit root node actually the

## Issuing command line queries

Using https://github.com/Gavinpub/ws (which supports subprotocol negotiation)

    ws -s graphql-ws ws://35.203.81.181:8080/graphql

    {"id":"3","type":"connection_init","payload":{"Authorization":"Bearer ......"},"variables":null}}

    {"id":"3","type":"start","payload":{"query":"subscription{searchTransactionsForward(query: \"status:executed\", lowBlockNum:0, highBlockNum:0){ trace{ block { num } id }}}","variables":null}}

    {"id":"3","type":"start","payload":{"query":"subscription{searchTransactionsForward(query: \"status:executed\", lowBlockNum:0, highBlockNum:0){ trace{ block { num } id }}}","variables":null}}

Using curl:
    curl http://localhost:8080/graphql -XPOST -d '{"query": "{ searchTransactionsForward(limit: 10, query: \"status:executed\") { cursor } }"}' -H "Authorization: Bearer $DFUSE" | jq .
    curl http://localhost:8080/graphql -XPOST -d '{"query": "{ blockIDByTime(time: \"2019-01-01T00:00:00Z\") { time num } }"}' -H "Authorization: Bearer $DFUSE" | jq .
