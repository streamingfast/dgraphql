// Copyright 2019 dfuse Platform Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"github.com/dfuse-io/dmetrics"
)

var perGraphqlLabels = []string{"type"}

// MetricSet contains all recorded metrics for this package
var MetricSet = dmetrics.NewSet()

// HeadBlockNumber records the latest block number seen by a pipeline (if present)
var HeadBlockNumber = MetricSet.NewHeadBlockNumber("dgraphql")

// HeadTimeDrift records the amount of time difference between 'now' and the block time of the latest block seen by a pipeline (if present)
var HeadTimeDrift = MetricSet.NewHeadTimeDrift("dgraphql")

// TotalRequestCount records total count of request (either a Query, Mutation or Subscription)
var TotalRequestCount = MetricSet.NewGauge("dgraphql_total_request_count", "total count of request served since startup time")

// InflightRequestCount records active count of request (either an active Query, Mutation or Subscription)
var InflightRequestCount = MetricSet.NewGauge("dgraphql_inflight_request_count", "inflight request count currently active")

// InflightQueryCount records active count of query being resolved, per GraphQL type
var InflightQueryCount = MetricSet.NewGaugeVec("dgraphql_inflight_query_count", perGraphqlLabels, "inflight query count currently active, per GraphQL type")

// TotalQueryCount records total count of query performed so far, per GraphQL type
var TotalQueryCount = MetricSet.NewGaugeVec("dgraphql_total_query_count", perGraphqlLabels, "total count of query started since startup time, per GraphQL type")

// QueryResponseTimes records histogram of "query" time taken to answer
var QueryResponseTimes = MetricSet.NewHistogramVec("dgraphql_query_response_times", perGraphqlLabels, "query response times histogram for percentile sampling, per GraphQL type")

// InflightMutationCount records active count of mutation being resolved, per GraphQL type
var InflightMutationCount = MetricSet.NewGaugeVec("dgraphql_inflight_mutation_count", perGraphqlLabels, "inflight mutation count currently active, per GraphQL type")

// TotalMutationCount records total count of mutation performed so far, per GraphQL type
var TotalMutationCount = MetricSet.NewGaugeVec("dgraphql_total_mutation_count", perGraphqlLabels, "total count of mutation started since startup time, per GraphQL type")

// MutationResponseTimes records histogram of "mutation" time taken to answer
var MutationResponseTimes = MetricSet.NewHistogramVec("dgraphql_mutation_response_times", perGraphqlLabels, "mutation response times histogram for percentile sampling, per GraphQL type")

// InflightSubscriptionCount records active count of subscription, per GraphQL type (top-level subscription element)
var InflightSubscriptionCount = MetricSet.NewGaugeVec("dgraphql_inflight_subscription_count", perGraphqlLabels, "inflight subscription count currently active, per GraphQL type")

// TotalSubscriptionCount records total count of query performed so far, per GraphQL type (top-level subscription element)
var TotalSubscriptionCount = MetricSet.NewGaugeVec("dgraphql_total_subscription_count", perGraphqlLabels, "total count of subscription started since startup time, per GraphQL type")
