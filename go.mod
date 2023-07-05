module github.com/streamingfast/dgraphql

go 1.13

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.10
	github.com/GeertJohan/go.rice v1.0.0
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.1
	github.com/graph-gophers/graphql-go v0.0.0-20191115155744-f33e81362277
	github.com/pkg/errors v0.9.1
	github.com/streamingfast/dauth v0.0.0-20210812020920-1c83ba29add1
	github.com/streamingfast/derr v0.0.0-20220301163149-de09cb18fc70
	github.com/streamingfast/dgrpc v0.0.0-20220301153539-536adf71b594
	github.com/streamingfast/dipp v1.0.1-0.20210811200841-d2cca4e058e6
	github.com/streamingfast/dmetering v0.0.0-20220301165106-a642bb6a21bd
	github.com/streamingfast/dmetrics v0.0.0-20210811180524-8494aeb34447
	github.com/streamingfast/dtracing v0.0.0-20220305214756-b5c0e8699839
	github.com/streamingfast/jsonpb v0.0.0-20210811021341-3670f0aa02d0
	github.com/streamingfast/logging v0.0.0-20220304214715-bc750a74b424
	github.com/streamingfast/opaque v0.0.0-20210811180740-0c01d37ea308
	github.com/streamingfast/pbgo v0.0.6-0.20220228185940-1bbaafec7d8a
	github.com/streamingfast/shutter v1.5.0
	github.com/stretchr/testify v1.8.1
	go.opencensus.io v0.24.0
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.53.0
	gotest.tools v2.2.0+incompatible
)

replace github.com/graph-gophers/graphql-go => github.com/streamingfast/graphql-go v0.0.0-20210204202750-0e485a040a3c
