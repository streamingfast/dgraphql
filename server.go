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
	"strings"
	"time"

	"github.com/dfuse-io/shutter"
	"github.com/streamingfast/dauth/authenticator"
	dauthAuthenticator "github.com/streamingfast/dauth/authenticator"
	_ "github.com/streamingfast/dauth/authenticator/null" // register plugin
	"github.com/streamingfast/dgraphql/static"
	"github.com/streamingfast/dmetering"
	"go.uber.org/zap"
)

type Server struct {
	*shutter.Shutter

	grpcListenAddr  string
	grpcSSL         bool
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

	grpcSSL := false
	if authenticator.GetAuthTokenRequirement() == dauthAuthenticator.AuthTokenRequired {
		grpcSSL = true // auth always requires SSL
	}
	if strings.Contains(grpcListenAddr, "*") {
		grpcSSL = true
		grpcListenAddr = strings.Replace(grpcListenAddr, "*", "", 1)
	}

	return &Server{
		Shutter:                  shutter.New(),
		grpcListenAddr:           grpcListenAddr,
		grpcSSL:                  grpcSSL,
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

	<-s.Shutter.Terminating()
	select {
	case <-s.Shutter.Terminated():
	case <-time.After(2 * time.Second): // it should take at most 1sec, usually way faster
		zlog.Warn("dgraphql not terminated gracefully")
	}

	if err := s.Err(); err != nil {
		zlog.Error("dgraphql terminated with error", zap.Error(err))
	} else {
		zlog.Info("dgraphql terminated")
	}
}
