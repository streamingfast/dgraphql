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
	"net/http"
	"time"

	dauthAuthenticator "github.com/dfuse-io/dauth/authenticator"
	"github.com/dfuse-io/derr"
	"github.com/dfuse-io/dgraphql"
	"github.com/dfuse-io/dgraphql/metrics"
	"github.com/dfuse-io/dmetering"
	"github.com/dfuse-io/dmetrics"
	"github.com/dfuse-io/shutter"
	"go.uber.org/zap"
)

type Config struct {
	HTTPListenAddr           string
	GRPCListenAddr           string
	AuthPlugin               string
	MeteringPlugin           string
	NetworkID                string
	OverrideTraceID          bool
	Protocol                 string
	JwtIssuerURL             string
	ApiKey                   string
	Schemas                  *dgraphql.Schemas
	DataIntegrityProofSecret string
}

type App struct {
	*shutter.Shutter
	config *Config
}

func New(config *Config) *App {
	return &App{
		Shutter: shutter.New(),
		config:  config,
	}
}

func (a *App) Run() error {
	zlog.Info("starting dgraphql eosio", zap.Reflect("config", a.config))

	dmetrics.Register(metrics.MetricSet)

	auth, err := dauthAuthenticator.New(a.config.AuthPlugin)
	derr.Check("unable to initialize dauth", err)

	meter, err := dmetering.New(a.config.MeteringPlugin)
	derr.Check("unable to initialize dmetering", err)
	dmetering.SetDefaultMeter(meter)

	zlog.Info("starting dgraphql server")
	server := dgraphql.NewServer(
		a.config.GRPCListenAddr,
		a.config.HTTPListenAddr,
		a.config.Protocol,
		a.config.NetworkID,
		a.config.OverrideTraceID,
		auth,
		meter,
		a.config.Schemas,
		a.config.DataIntegrityProofSecret,
		a.config.JwtIssuerURL,
		a.config.ApiKey,
	)

	a.OnTerminating(server.Shutdown)
	server.OnTerminated(a.Shutdown)

	go server.Launch()

	return nil
}

func (a *App) IsReady() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	url := fmt.Sprintf("http://%s/healthz", a.config.HTTPListenAddr)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		zlog.Warn("IsReady request building error", zap.Error(err))
		return false
	}
	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		zlog.Debug("IsReady request execution error", zap.Error(err))
		return false
	}

	if res.StatusCode == 200 {
		return true
	}
	return false
}
