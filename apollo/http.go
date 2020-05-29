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

package apollo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dfuse-io/dauth/authenticator"
	"github.com/dfuse-io/dtracing"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const protocolGraphQLWS = "graphql-ws"

var upgrader = websocket.Upgrader{
	CheckOrigin:  func(r *http.Request) bool { return true },
	Subprotocols: []string{protocolGraphQLWS},
}

type Middleware struct {
	service          GraphQLService
	authenticateFunc AuthenticateFunc
}

func NewMiddleware(service GraphQLService, authenticate func(context.Context, string, string) (context.Context, error)) *Middleware {
	return &Middleware{
		service: service,
		authenticateFunc: func(ctx context.Context, r *http.Request, payload map[string]interface{}) (context.Context, error) {
			if tokenObject, ok := payload["Authorization"]; ok {
				if tokenString, ok := tokenObject.(string); ok {
					tokenString := strings.TrimPrefix(tokenString, "Bearer ")

					ip := authenticator.RealIPFromRequest(r)
					ctx, err := authenticate(ctx, tokenString, ip)
					if err != nil {
						return nil, err
					}
					return ctx, nil
				}
				return nil, fmt.Errorf("expected 'Authorization' to be of string type")
			}

			return nil, fmt.Errorf("missing 'Authorization' from 'connection_init' payload")
		},
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionProtocols := websocket.Subprotocols(r)
		for _, subprotocol := range connectionProtocols {
			if subprotocol == protocolGraphQLWS {
				ws, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					zlog.Debug("unable to upgrade HTTP connection", zap.Error(err))
					return
				}

				if ws.Subprotocol() != protocolGraphQLWS {
					zlog.Debug("created websocket connection is not using right subprotocol",
						zap.String("expected_protocol", protocolGraphQLWS),
						zap.String("actual_protocol", ws.Subprotocol()),
					)

					ws.Close()
					return
				}

				connectionTraceID := dtracing.GetTraceID(r.Context())

				zlog.Debug("websocket connection initialized correctly, continuing connection process")
				go Connect(connectionTraceID.String(), ws, m.service, Authentication(r, m.authenticateFunc))
				return
			}
		}

		zlog.Debug("this connection didn't had expected protocol, assuming it's a normal HTTP connection",
			zap.String("expected_protocol", protocolGraphQLWS),
			zap.Strings("received_protocols", connectionProtocols),
		)

		next.ServeHTTP(w, r)
	})
}
