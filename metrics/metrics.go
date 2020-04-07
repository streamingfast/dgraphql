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

var metrics = dmetrics.NewSet()

var QueryResponseTimes = metrics.NewHistogram("query_response_times", "query response times histogram for percentile sampling")
var InflightQueryCount = metrics.NewGauge("inflight_query_count", "inflight query count currently active")
var InflightSubscriptionCount = metrics.NewGauge("inflight_subscription_count", "inflight subscription count currently active")
var HeadBlockNumber = metrics.NewHeadBlockNumber("dgraphql")
var HeadTimeDrift = metrics.NewHeadTimeDrift("dgraphql")

func ServeMetrics() {

	dmetrics.Serve(":9102")
}

func init() {
	metrics.Register()
}
