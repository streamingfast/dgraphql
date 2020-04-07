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

	"github.com/dfuse-io/derr"
	"github.com/dfuse-io/dtracing"
	"github.com/graph-gophers/graphql-go/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnwrapError(ctx context.Context, err error) *errors.QueryError {
	if se, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
		sts := se.GRPCStatus().Proto()
		msg := sts.Message
		if traceID := dtracing.GetTraceID(ctx).String(); traceID != "" {
			msg = fmt.Sprintf("%s (trace_id: %s)", msg, traceID)
		}
		return &errors.QueryError{
			Message: msg,
			Extensions: map[string]interface{}{
				"code":     codes.Code(sts.Code).String(),
				"terminal": true,
			},
		}
	}

	msg := err.Error()
	if traceID := dtracing.GetTraceID(ctx).String(); traceID != "" {
		msg = fmt.Sprintf("%s (trace_id: %s)", msg, traceID)
	}
	return &errors.QueryError{
		Message: msg,
		Extensions: map[string]interface{}{
			"code":     "Internal",
			"terminal": true,
		},
	}
}

func Errorf(ctx context.Context, format string, args ...interface{}) *errors.QueryError {
	return UnwrapError(ctx, fmt.Errorf(format, args...))
}

func Status(ctx context.Context, code codes.Code, message string) *errors.QueryError {
	return UnwrapError(ctx, derr.Status(code, message))
}
