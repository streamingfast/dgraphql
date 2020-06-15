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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/dfuse-io/dauth/authenticator"
	"github.com/dfuse-io/dgraphql/analytics"
	"github.com/dfuse-io/dgraphql/insecure"
	"github.com/dfuse-io/dgrpc"
	"github.com/dfuse-io/dmetering"
	"github.com/dfuse-io/jsonpb"
	"github.com/dfuse-io/logging"
	pbgraphql "github.com/dfuse-io/pbgo/dfuse/graphql/v1"
	pbstruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/errors"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) startGRPCServer() {

	if !s.authenticator.IsAuthenticationTokenRequired() {
		s.startGRPCServerInsecure()
		return
	}

	internalGrpcServer := newGRPCServer(s.schemas.GetSchema(WithAlpha()), s.authenticator, s.overrideTraceID)

	grpcRouter := mux.NewRouter()
	grpcRouter.Path("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	grpcRouter.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = maybeRewriteLegacyPaths(r.URL.Path)
		internalGrpcServer.ServeHTTP(w, r)
	})

	grpcListener, err := net.Listen("tcp", s.grpcListenAddr)
	if err != nil {
		s.Shutter.Shutdown(fmt.Errorf("failed listening grpc %q: %w", s.grpcListenAddr, err))
		return
	}

	errorLogger, err := zap.NewStdLogAt(zlog, zap.ErrorLevel)
	if err != nil {
		s.Shutter.Shutdown(fmt.Errorf("unable to create error logger: %w", err))
		return
	}

	grpcServer := http.Server{
		Handler: grpcRouter,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{insecure.Cert},
			ClientCAs:    insecure.CertPool,
			ClientAuth:   tls.VerifyClientCertIfGiven,
		},
		ErrorLog: errorLogger,
	}

	go func() {
		zlog.Info("serving gRPC", zap.String("grpc_addr", s.grpcListenAddr))
		if err := grpcServer.ServeTLS(grpcListener, "", ""); err != nil {
			s.Shutter.Shutdown(fmt.Errorf("error on gs.Serve: %w", err))
		}
	}()
}

func (s *Server) startGRPCServerInsecure() {
	grpcListener, err := net.Listen("tcp", s.grpcListenAddr)
	if err != nil {
		s.Shutter.Shutdown(fmt.Errorf("failed listening grpc %q: %w", s.grpcListenAddr, err))
		return
	}

	gs := newGRPCServer(s.schemas.GetSchema(WithAlpha()), s.authenticator, s.overrideTraceID)
	go func() {
		zlog.Info("serving gRPC", zap.String("grpc_addr", s.grpcListenAddr))
		if err := gs.Serve(grpcListener); err != nil {
			s.Shutter.Shutdown(fmt.Errorf("error on gs.Serve: %w", err))
		}
	}()
}

func newGRPCServer(schema *graphql.Schema, authenticator authenticator.Authenticator, overrideTraceID bool) *grpc.Server {
	serverOptions := []dgrpc.ServerOption{dgrpc.WithLogger(zlog)}
	if overrideTraceID {
		serverOptions = append(serverOptions, dgrpc.OverrideTraceID())
	}

	zlog.Info("configuring grpc server")
	gs := dgrpc.NewServer(serverOptions...)
	pbgraphql.RegisterGraphQLServer(gs, NewEndpointServer(schema, authenticator))

	return gs
}

type EndpointServer struct {
	schema        *graphql.Schema
	authenticator authenticator.Authenticator
}

func NewEndpointServer(schema *graphql.Schema, authenticator authenticator.Authenticator) *EndpointServer {
	return &EndpointServer{
		schema:        schema,
		authenticator: authenticator,
	}
}

func (s *EndpointServer) Execute(req *pbgraphql.Request, stream pbgraphql.GraphQL_ExecuteServer) error {
	ctx, span := trace.StartSpan(stream.Context(), "executing_request")
	defer span.End()

	zlogger := logging.Logger(ctx, zlog)
	zlogger.Debug("executing request")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err := status.Errorf(codes.Unauthenticated, "missing metadata")
		analytics.TrackSubscriptionError(ctx, "grpc", err)
		return err
	}

	token := ""
	if s.authenticator.IsAuthenticationTokenRequired() {
		authValues := md["authorization"]
		if len(authValues) < 1 {
			err := status.Errorf(codes.Unauthenticated, "missing 'authorization' metadata field")
			return err
		}
		token = strings.TrimPrefix(authValues[0], "Bearer ")
	}

	mdIn, _ := metadata.FromIncomingContext(ctx)
	xff := mdIn.Get("x-forwarded-for")
	ip := authenticator.RealIP(strings.Join(xff, ", "))

	var e error
	ctx, e = s.authenticator.Check(ctx, token, ip)
	if e != nil {
		return status.Errorf(codes.Unauthenticated, e.Error())
	}

	analytics.TrackSubscriptionStart(ctx, "grpc")

	//////////////////////////////////////////////////////////////////////
	// Billable event on GraphQL Subscriptions
	// WARNING : Here we only track Ingress bytes
	//////////////////////////////////////////////////////////////////////
	dmetering.EmitWithContext(dmetering.Event{
		Source:       "dgraphql",
		Kind:         "GraphQL Subscription",
		Method:       "", //TODO For now we will need by able to aggregate Ingress / Egress per method
		IngressBytes: int64(req.XXX_Size()),
	}, ctx)
	//////////////////////////////////////////////////////////////////////

	vars := decodeToMap(req.Variables)

	c, err := s.schema.Subscribe(ctx, req.Query, req.OperationName, vars)
	if err != nil {
		analytics.TrackSubscriptionError(ctx, "grpc", err)
		return err
	}

	for {
		select {
		case <-ctx.Done():
			analytics.TrackSubscriptionContextDone(ctx, "grpc")
			return nil
		case payload, more := <-c:
			if !more {
				analytics.TrackSubscriptionComplete(ctx, "grpc")
				return nil
			}

			if subErr, ok := payload.(errors.SubscriptionError); ok {
				if err := subErr.SubscriptionError(); err != nil {
					err2 := stream.Send(&pbgraphql.Response{
						Errors: convertToGRPCErrors(err),
					})
					if err2 != nil {
						return err2
					}
					analytics.TrackSubscriptionError(ctx, "grpc", err)
					return nil
				}
			}

			gqlResp := payload.(*graphql.Response)

			rpcResp := &pbgraphql.Response{
				Data:   string(gqlResp.Data),
				Errors: convertGQLToGRPCErrors(gqlResp.Errors),
			}

			//////////////////////////////////////////////////////////////////////
			// Billable event on GraphQL Subscriptions
			// WARNING : Here we only track Egress bytes
			//////////////////////////////////////////////////////////////////////
			dmetering.EmitWithContext(dmetering.Event{
				Source:      "dgraphql",
				Kind:        "GraphQL Subscription",
				Method:      "", //TODO For now we will need by able to aggregate Ingress / Egress per method
				EgressBytes: int64(rpcResp.XXX_Size()),
			}, ctx)
			//////////////////////////////////////////////////////////////////////

			if err := stream.Send(rpcResp); err != nil {
				analytics.TrackSubscriptionError(ctx, "grpc", err)
				return err
			}
		}
	}
}

func convertToGRPCErrors(errs ...error) (out []*pbgraphql.Error) {
	for _, err := range errs {
		if gqlError, ok := err.(*errors.QueryError); ok {
			out = append(out, convertToGRPCError(gqlError))
		} else {
			out = append(out, convertToGRPCError(&errors.QueryError{
				Message: err.Error(),
			}))
		}
	}
	return
}

func convertGQLToGRPCErrors(errs []*errors.QueryError) (out []*pbgraphql.Error) {
	for _, err := range errs {
		out = append(out, convertToGRPCError(err))
	}
	return
}

func convertToGRPCError(err *errors.QueryError) (out *pbgraphql.Error) {
	out = &pbgraphql.Error{
		Message: err.Message,
	}

	for _, loc := range err.Locations {
		out.Locations = append(out.Locations, &pbgraphql.SourceLocation{
			Line:   int32(loc.Line),
			Column: int32(loc.Column),
		})
	}

	if len(err.Path) != 0 {
		out.Path = &pbstruct.ListValue{}
	}
	for _, pathEl := range err.Path {
		switch pathVal := pathEl.(type) {
		case string:
			out.Path.Values = append(out.Path.Values, &pbstruct.Value{Kind: &pbstruct.Value_StringValue{pathVal}})
		case int:
			out.Path.Values = append(out.Path.Values, &pbstruct.Value{Kind: &pbstruct.Value_NumberValue{float64(pathVal)}})
		default:
			panic(fmt.Sprintf("unknown path segment type: %T", pathEl))
		}
	}

	if err.Extensions != nil {
		cnt, err2 := json.Marshal(err.Extensions)
		if err2 != nil {
			zlog.Error("failed json marshalling extensions in GraphQL error payload", zap.Error(err2))
			return
		}

		out.Extensions = &pbstruct.Struct{}
		if err2 := jsonpb.UnmarshalString(string(cnt), out.Extensions); err2 != nil {
			zlog.Error("failed json unmarshalling extensions in gRPC error payload", zap.Error(err2))
			return
		}
	}

	return
}

func decodeToMap(s *pbstruct.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	m := map[string]interface{}{}
	for k, v := range s.Fields {
		m[k] = decodeValue(v)
	}
	return m
}

func decodeValue(v *pbstruct.Value) interface{} {
	switch k := v.Kind.(type) {
	case *pbstruct.Value_NullValue:
		return nil
	case *pbstruct.Value_NumberValue:
		return k.NumberValue
	case *pbstruct.Value_StringValue:
		return k.StringValue
	case *pbstruct.Value_BoolValue:
		return k.BoolValue
	case *pbstruct.Value_StructValue:
		return decodeToMap(k.StructValue)
	case *pbstruct.Value_ListValue:
		s := make([]interface{}, len(k.ListValue.Values))
		for i, e := range k.ListValue.Values {
			s[i] = decodeValue(e)
		}
		return s
	default:
		panic("protostruct: unknown kind")
	}
}

func maybeRewriteLegacyPaths(urlPath string) string {
	if strings.HasPrefix(urlPath, "/dfuse.eosio.v1.GraphQL") {
		return strings.Replace(urlPath, "/dfuse.eosio.v1.GraphQL", "/dfuse.graphql.v1.GraphQL", 1)
	}
	return urlPath
}
