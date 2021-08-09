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

	"github.com/dfuse-io/logging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var traceEnabled = logging.IsTraceEnabled("dgraphql", "github.com/streamingfast/dgraphql")
var zlog *zap.Logger

func init() {
	logging.Register("github.com/streamingfast/dgraphql", &zlog)
}

type graphqlLogger struct{}

func (*graphqlLogger) LogPanic(ctx context.Context, value interface{}) {
	var err error
	if v, ok := value.(error); ok {
		err = v
	} else {
		err = fmt.Errorf("unknown error: %s", value)
	}

	logging.Logger(ctx, zlog).Error("graphlql resolver panicked", zap.Error(errors.WithStack(err)))
}
