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

package dgraphql

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"

	stackdriverPropagation "contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go"
	gerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/streamingfast/dtracing"
)

func CompressionMiddleware(next http.Handler) http.Handler {
	return handlers.CompressHandlerLevel(next, gzip.BestSpeed)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return dtracing.NewAddTraceIDAwareLoggerMiddleware(next, zlog, &stackdriverPropagation.HTTPFormat{})
}

func NewCORSMiddleware() mux.MiddlewareFunc {
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "X-Eos-Push-Guarantee"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "OPTIONS"})
	maxAge := handlers.MaxAge(86400) // 24 hours - hard capped by Firefox / Chrome is max 10 minutes

	return handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods, maxAge)
}

func AuthErrorHandler(w http.ResponseWriter, ctx context.Context, err error) {
	_ = json.NewEncoder(w).Encode(&graphql.Response{
		Errors: []*gerrors.QueryError{
			UnwrapError(ctx, err),
		},
	})
}
