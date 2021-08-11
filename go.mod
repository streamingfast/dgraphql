module github.com/streamingfast/dgraphql

go 1.13

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.8
	github.com/GeertJohan/go.rice v1.0.0
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/dfuse-io/dipp v1.0.1-0.20200407033930-5c17c531c3c4
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1
	github.com/graph-gophers/graphql-go v0.0.0-20191115155744-f33e81362277
	github.com/pkg/errors v0.9.1
	github.com/streamingfast/bstream v0.0.2-0.20210811181043-4c1920a7e3e3 // indirect
	github.com/streamingfast/dauth v0.0.0-20210811181149-e8fd545948cc
	github.com/streamingfast/derr v0.0.0-20210811180100-9138d738bcec
	github.com/streamingfast/dgrpc v0.0.0-20210811180351-8646818518b2
	github.com/streamingfast/dmetering v0.0.0-20210811181351-eef120cfb817
	github.com/streamingfast/dmetrics v0.0.0-20210811180524-8494aeb34447
	github.com/streamingfast/dtracing v0.0.0-20210811175635-d55665d3622a
	github.com/streamingfast/jsonpb v0.0.0-20210811021341-3670f0aa02d0
	github.com/streamingfast/logging v0.0.0-20210811175431-f3b44b61606a // indirect
	github.com/streamingfast/opaque v0.0.0-20210811180740-0c01d37ea308
	github.com/streamingfast/pbgo v0.0.6-0.20210811160400-7c146c2db8cc // indirect
	github.com/streamingfast/shutter v1.5.0
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.23.0
	go.uber.org/atomic v1.6.0
	go.uber.org/zap v1.15.0
	google.golang.org/grpc v1.39.1
	gotest.tools v2.2.0+incompatible
)

replace github.com/graph-gophers/graphql-go => github.com/dfuse-io/graphql-go v0.0.0-20210204202750-0e485a040a3c

// This is required to fix build where 0.1.0 version is not considered a valid version because a v0 line does not exists
// We replace with same commit, simply tricking go and tell him that's it's actually version 0.0.3
replace github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8

replace github.com/blendle/zapdriver => github.com/blendle/zapdriver v1.3.1
