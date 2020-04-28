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
	"github.com/dfuse-io/dauth"
	_ "github.com/dfuse-io/dauth/null" // register plugin
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
	authenticator   dauth.Authenticator
	metering        dmetering.Metering
	schemas         *Schemas

	DataIntegrityProofSecret string
	jwtIssuerURL             string
	apiKey                   string
}

func NewServer(
	grpcListenAddr string,
	httpListenAddr string,
	protocol string,
	networkID string,
	overrideTraceID bool,
	authenticator dauth.Authenticator,
	metering dmetering.Metering,
	schemas *Schemas,
	dataIntegrityProofSecret string,
	jwtIssuerURL string,
	apiKey string,

) *Server {
	if authenticator == nil {
		authenticator, _ = dauth.New("null://")
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
