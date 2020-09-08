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
	"github.com/dfuse-io/dauth/authenticator"
	dauthAuthenticator "github.com/dfuse-io/dauth/authenticator"
	_ "github.com/dfuse-io/dauth/authenticator/null" // register plugin
	"github.com/dfuse-io/dgraphql/static"
	"github.com/dfuse-io/dmetering"
	"github.com/dfuse-io/shutter"
	"go.uber.org/zap"
)

type Server struct {
	*shutter.Shutter

	grpcListenAddr  string
	httpListenAddr  string
	protocol        string
	networkID       string
	overrideTraceID bool
	authenticator   authenticator.Authenticator
	metering        dmetering.Metering
	schemas         *Schemas

	DataIntegrityProofSecret string
	jwtIssuerURL             string
	apiKey                   string
	predfinedGraphqlExamples []*static.GraphqlExample
}

func NewServer(
	grpcListenAddr string,
	httpListenAddr string,
	protocol string,
	networkID string,
	overrideTraceID bool,
	authenticator authenticator.Authenticator,
	metering dmetering.Metering,
	schemas *Schemas,
	dataIntegrityProofSecret string,
	jwtIssuerURL string,
	apiKey string,
	predfinedGraphqlExamples []*static.GraphqlExample,

) *Server {
	if authenticator == nil {
		authenticator, _ = dauthAuthenticator.New("null://")
	}
	return &Server{
		Shutter:                  shutter.New(),
		grpcListenAddr:           grpcListenAddr,
		httpListenAddr:           httpListenAddr,
		protocol:                 protocol,
		networkID:                networkID,
		overrideTraceID:          overrideTraceID,
		authenticator:            authenticator,
		metering:                 metering,
		schemas:                  schemas,
		DataIntegrityProofSecret: dataIntegrityProofSecret,
		jwtIssuerURL:             jwtIssuerURL,
		apiKey:                   apiKey,
		predfinedGraphqlExamples: predfinedGraphqlExamples,
	}
}

func (s *Server) Launch() {
	s.startHTTPServer()
	s.startGRPCServer()

	select {
	case <-s.Shutter.Terminating():
		if err := s.Err(); err != nil {
			zlog.Error("dgraphql terminated with error", zap.Error(err))
		} else {
			zlog.Info("dgraphql terminated")
		}
	}
}
