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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dfuse-io/dgraphql/metrics"
	"github.com/dfuse-io/dtracing"
	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/introspection"
	gqtrace "github.com/graph-gophers/graphql-go/trace"
	"go.opencensus.io/trace"
)

type OpencensusTracer struct {
}

func (t *OpencensusTracer) TraceQuery(
	ctx context.Context,
	queryString string,
	operationName string,
	variables map[string]interface{},
	varTypes map[string]*introspection.Type,
) (context.Context, gqtrace.TraceQueryFinishFunc) {
	startTime := time.Now()
	metrics.InflightQueryCount.Inc()

	spanCtx, span := dtracing.StartSpan(ctx, "GraphQL request",
		"graphql.query", queryString,
		"graphql.operationName", operationName,
		"graphql.variablesCount", len(variables),
	)

	return spanCtx, func(errs []*errors.QueryError) {
		if len(errs) > 0 {
			setSpanError(span, errs)
		}

		span.End()
		metrics.InflightQueryCount.Dec()
		metrics.QueryResponseTimes.ObserveSince(startTime)
	}
}

func (t *OpencensusTracer) TraceField(
	ctx context.Context,
	label, typeName, fieldName string,
	trivial bool,
	args map[string]interface{},
) (context.Context, gqtrace.TraceFieldFinishFunc) {
	if trivial {
		return ctx, noop
	}

	var i = 0
	keyedAttributes := make([]interface{}, len(args)*2)
	for name, value := range args {
		keyedAttributes[i] = "graphql.args." + name
		i++
		keyedAttributes[i] = sanitizeTraceAttributeValue(value)
		i++
	}

	keyedAttributes = append(keyedAttributes,
		"graphql.type", typeName,
		"graphql.field", fieldName,
	)

	spanCtx, span := dtracing.StartSpan(ctx, "GraphQL request", keyedAttributes...)

	return spanCtx, func(err *errors.QueryError) {
		if err != nil {
			setSpanError(span, []*errors.QueryError{err})
		}

		span.End()
	}
}

func sanitizeTraceAttributeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case bool, int, int8, int16, int32, int64, uintptr, uint, uint8, uint16, uint32, uint64:
		return v
	case string, fmt.Stringer:
		return v
	case map[string]interface{}:
		pairs := make([]string, len(v))

		i := 0
		for k, v := range v {
			pairs[i] = fmt.Sprintf("%s: %s", k, v)
			i++
		}

		return strings.Join(pairs, ", ")
	default:
		return fmt.Sprintf("%T", v)
	}
}

func setSpanError(span *trace.Span, errs []*errors.QueryError) {
	if len(errs) > 0 {
		msg := errs[0].Error()
		if len(errs) > 1 {
			msg += fmt.Sprintf(" (and %d more errors)", len(errs)-1)
		}

		// What's the correct way to track error in the span?
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown,
			Message: msg,
		})

		span.AddAttributes(trace.StringAttribute("graphql.error", msg))
	}
}

func noop(*errors.QueryError) {}
