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

package static

//go:generate rice embed-go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const IndexFilename = "graphiql.html"
const (
	MIME_TYPE_CSS  = "text/css"
	MIME_TYPE_HTML = "text/html"
	MIME_TYPE_JS   = "application/javascript"
	MIME_TYPE_JSON = "application/json"
	MIME_TYPE_SVG  = "image/svg+xml"
)

// RegisterStaticRoutes registers all GraphiQL static route enabling traffic of `/graphiql` paths.
// The `predfinedGraphqlExamples` are used to prefill the GraphQL history for the specified protocol
// and network.
//
// The `jwtIssuerURL` is used to configure the `client-js` library so we are able to easily fetch
// authentication tokens.
func RegisterStaticRoutes(router *mux.Router, protocol, network, apiKey, jwtIssuerURL string, predfinedGraphqlExamples []*GraphqlExample) error {
	zlog.Info("registering static route")
	box := rice.MustFindBox("dgraphql-build")

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/graphiql/", 302)
	})

	router.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/graphiql/", 302)
	})

	serveIndexHTML(router, box, "/graphiql/", apiKey, jwtIssuerURL)
	serveFileAsset(router, box, "/graphiql/graphiql_dfuse_override.css", "graphiql_dfuse_override.css", MIME_TYPE_CSS)
	serveFileAsset(router, box, "/graphiql/helper.js", "helper.js", MIME_TYPE_JS)
	serveFileAsset(router, box, "/graphiql/dfuse_logo.svg", "dfuse_logo.svg", MIME_TYPE_SVG)

	graphqlExamplesJSON, err := json.Marshal(predfinedGraphqlExamples)
	if err != nil {
		return fmt.Errorf("marshal graphql examples: %w", err)
	}
	serveInMemoryAsset(router, "/graphiql/predefined_examples.json", graphqlExamplesJSON, MIME_TYPE_JSON)

	configJSON := []byte(fmt.Sprintf(`{"protocol":"%s","network":"%s"}`, protocol, network))
	serveInMemoryAsset(router, "/graphiql/config.json", configJSON, MIME_TYPE_JSON)

	// Redirects since it was supported at some point, redirects everyone to `GraphiQL` instead
	router.HandleFunc("/playground", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/graphiql/", 302)
	})

	router.HandleFunc("/playground/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/graphiql/", 302)
	})

	return nil
}

func serveInMemoryAsset(router *mux.Router, path string, content []byte, contentType string) {
	router.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write(content)
	}))
}

func serveIndexHTML(router *mux.Router, box *rice.Box, path string, apiKey string, jwtIssuerURL string) {
	zlog.Info("setting up index http handler")
	router.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		zlog.Info("serving index",
			zap.String("path", path),
			zap.String("jwt_issuer_url", jwtIssuerURL),
			zap.String("api_key", apiKey),
			zap.String("file_to_template", IndexFilename),
		)
		reader, err := templatedIndex(box, apiKey, jwtIssuerURL)
		if err != nil {
			zlog.Error("unable to serve graphiql.html", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to read asset"))
			return
		}

		w.Header().Set("Content-Type", MIME_TYPE_HTML)
		io.Copy(w, reader)
	}))
}

func serveFileAsset(router *mux.Router, box *rice.Box, path string, asset string, contentType string, options ...interface{}) {
	router.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reader, err := box.Open(asset)
		if err != nil {
			zlog.Error("unable to server static file asset from box", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unable to read asset"))
			return
		}
		defer reader.Close()

		w.Header().Set("Content-Type", contentType)

		// We ignore the error since if we are unable to write to HTTP pipe, it's probably broken
		io.Copy(w, reader)
	}))
}

func templatedIndex(box *rice.Box, apiKey string, jwtIssuerURL string) (*bytes.Reader, error) {
	zlog.Info("rendering templated index")
	indexContent, err := box.Bytes(IndexFilename)
	if err != nil {
		return nil, err
	}

	config := map[string]interface{}{
		"apiKey":  apiKey,
		"authUrl": jwtIssuerURL,
	}

	tpl, err := template.New(IndexFilename).Funcs(template.FuncMap{
		"json": func(v interface{}) (template.JS, error) {
			cnt, err := json.Marshal(v)
			return template.JS(cnt), err
		},
	}).Delims("--==", "==--").Parse(string(indexContent))
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, config); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}
