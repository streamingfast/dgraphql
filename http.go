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
	"fmt"
	"net"
	"net/http"
	"strings"

	dauthMiddleware "github.com/dfuse-io/dauth/authenticator/middleware"
	"github.com/dfuse-io/dmetering"

	"github.com/dfuse-io/derr"
	"github.com/dfuse-io/dgraphql/apollo"
	"github.com/dfuse-io/dgraphql/static"
	"github.com/dfuse-io/dipp"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go/relay"
	"go.uber.org/zap"
)

func (s *Server) startHTTPServer() {
	apolloAuthenticator := func(ctx context.Context, token string, ip string) (context.Context, error) {
		return s.authenticator.Check(ctx, token, ip)
	}

	router := mux.NewRouter()
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if derr.IsShuttingDown() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	staticRouter := router.PathPrefix("/").Subrouter()
	static.RegisterStaticRoutes(staticRouter, s.protocol, s.getNetworkName(), s.apiKey, s.jwtIssuerURL)

	restRouter := router.PathPrefix("/").Subrouter()
	restRouter.Use(LoggingMiddleware)
	restRouter.Use(apollo.NewMiddleware(s.schemas.GetSchema(WithAlpha()), apolloAuthenticator).Handler)
	restRouter.Use(dauthMiddleware.NewAuthMiddleware(s.authenticator, AuthErrorHandler).Handler)
	if s.DataIntegrityProofSecret != "" {
		restRouter.Use(dipp.NewProofMiddlewareFunc(s.DataIntegrityProofSecret))
	}

	//////////////////////////////////////////////////////////////////////
	// Billable event on GraphQL Query
	// WARNING: Middleware is **configured** to ONLY track Query Ingress / Egress bytes.
	//          This means that the middleware DOES NOT track Query requests / responses.
	//          User-Query-level Req / Resp (Docs) is counted in the different Resolvers
	//////////////////////////////////////////////////////////////////////
	restRouter.Use(dmetering.NewMeteringMiddlewareFuncWithOptions(s.metering, "dgraphql", "GraphQL Query", false, true))
	//////////////////////////////////////////////////////////////////////

	// For now, we always serve the full schema, no matter what
	standardSchemaHandler := &relay.Handler{Schema: s.schemas.GetSchema(WithAlpha())}

	restRouter.Handle("/graphql", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isRequestingAlphaSchema(r) {
			standardSchemaHandler.ServeHTTP(w, r)
		} else {
			standardSchemaHandler.ServeHTTP(w, r)
		}
	}))

	// http
	httpListener, err := net.Listen("tcp", s.httpListenAddr)
	if err != nil {
		s.Shutter.Shutdown(fmt.Errorf("failed listening http %q: %w", s.httpListenAddr, err))
		return

	}

	corsMiddleware := NewCORSMiddleware()
	httpServer := http.Server{
		Handler: handlers.CompressHandlerLevel(corsMiddleware(router), gzip.BestSpeed),
	}

	go func() {
		zlog.Info("serving HTTP", zap.String("http_addr", s.httpListenAddr))
		if err := httpServer.Serve(httpListener); err != nil {
			s.Shutter.Shutdown(fmt.Errorf("error on http.Serve: %w", err))
		}
	}()

	return
}

func (s *Server) getNetworkName() string {
	parts := strings.SplitN(s.networkID, "-", 2)
	if len(parts) >= 2 {
		return parts[1]
	}

	return ""
}

func isRequestingAlphaSchema(r *http.Request) bool {
	return r.FormValue("alpha-schema") == "true" || r.FormValue("alphaSchema") == "true" || hasHeader("x-alpha-schema", "true", r.Header)
}

func hasHeader(name, value string, headers http.Header) bool {
	values, found := headers[name]
	if !found {
		return false
	}

	for _, actualValue := range values {
		if actualValue == value {
			return true
		}
	}

	return false
}
