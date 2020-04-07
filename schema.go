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
	"bytes"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"go.uber.org/zap"
)


var SchemaRegistry = map[string][]byte{}

func RegisterSchema(schemaPrefix, schemaName string, schemaData []byte) {
	zlog.Debug("registering schema",
		zap.String("schema_prefix", schemaPrefix),
	zap.String("schema_name", schemaName),
	)
	schemaFullName := schemaPrefix + schemaName
	if _, found := SchemaRegistry[schemaFullName]; found {
		panic(fmt.Sprintf("Cannot register a graphql schema with the same full name: %s", schemaFullName))
	}
	SchemaRegistry[schemaFullName] = schemaData
}


type schemaID string

const (
	standard schemaID = "standard"
	alpha    schemaID = "alpha"
)

type SchemaSelector interface {
	SchemaID() schemaID
	String() string
}

func WithAlpha() SchemaSelector {
	return withAlpha(alpha)
}

type Schemas struct {
	standard *graphql.Schema
	alpha    *graphql.Schema

	commonRawSchema *string
	alphaRawSchema  *string
}


func collectAssetData(isIncluded func(name string) bool) (schemasData [][]byte) {
	for name, data := range SchemaRegistry {
		if isIncluded(name) {
			schemasData = append(schemasData, data)
		}
	}
	return schemasData
}

func NewSchemas(resolver interface{}) (*Schemas, error) {

	commonRawSchema := buildSchemaString("common", collectAssetData(isCommonAssetName))
	alphaRawSchema := buildSchemaString("alpha", collectAssetData(isCommonAssetName))

	if commonRawSchema == nil {
		panic(fmt.Errorf("it's invalid to have a nil common raw schema"))
	}

	schemas := &Schemas{
		commonRawSchema: commonRawSchema,
		alphaRawSchema:  alphaRawSchema,
	}

	var err error
	schemas.standard, err = parseSchema("standard", resolver, schemas.getStandardString())
	if err != nil {
		// We return schemas even in case of error for debugging purposes (getting the raw Schema string)
		return schemas, fmt.Errorf("unable to build standard schema: %s", err)
	}

	if alphaRawSchema != nil {
		schemas.alpha, err = parseSchema("alpha", resolver, schemas.getAlphaString())
		if err != nil {
			// We return schemas even in case of error for debugging purposes (getting the raw Schema string)
			return schemas, fmt.Errorf("unable to build alpha schema: %s", err)
		}
	}

	return schemas, nil
}

func (s *Schemas) GetSchema(selectors ...SchemaSelector) *graphql.Schema {
	zlog.Info("requesting schema", zap.Reflect("selectors", selectors))
	if isSelectedSchema(alpha, selectors) && s.alpha != nil {
		zlog.Info("returning alpha schema")
		return s.alpha
	}

	zlog.Info("returning standard schema")
	return s.standard
}

func (s *Schemas) getStandardString() string {
	// We still use merge schemas here because the method add some comments to merged types
	return MergeSchemas("", *s.commonRawSchema)
}

var typeQueryOrSubscriptionRegexp = regexp.MustCompile(`(?s)type\s+(Query|Subscription)\s*\{([^}]*)\}`)

func (s *Schemas) getAlphaString() string {
	if s.alphaRawSchema == nil {
		return ""
	}

	return MergeSchemas(*s.alphaRawSchema, *s.commonRawSchema)
}

// MergeSchemas is a poor man way of merging some specific GraphQL types together
// to form a single schemas.
//
// The algorithm simply reads the schema and look for the pattern
// `type Query {(.*)}` and `type Subscription {(.*)}`, collecting them
// along the way and removing from the actual buffer.
//
// Once the full schema has been processed, we simply append a `type Query { ... }`
// node with the content (regexp sub-match) of the various collected queries in
// the first pass, same for `Subscription`.
//
// This creates a final `Query` and `Subscription` type of the collected merged
// elements.
//
// **Important** This is regexp based, so the current drawback of this technique
//               is that the inner sub-match content must not contain a `{ }` pair.
//               This is easy to achieve and maintain.
func MergeSchemas(first string, second string) string {

	full := first + "\n" + second
	matches := typeQueryOrSubscriptionRegexp.FindAllStringSubmatchIndex(full, -1)

	zlog.Debug("find query and subscription graphql types for schemas merging", zap.Int("match_count", len(matches)))
	if len(matches) == 0 {
		return full
	}

	var querySections []string
	var subscriptionSections []string

	// To be more compact, each match is a series of pairs,
	// so if there is 4 elements in the match, it means there
	// is the full match (0) and the first capturing group
	// match (1). The full start/end of full match is `match[0]/match[1]`
	// and the first capturing group is `match[2]/match[3]`
	for i, match := range matches {
		zlog.Debug("schema merging section match", zap.Reflect("match", match), zap.Int("index", i))

		if len(match) != 6 {
			panic(fmt.Errorf("expecting the match array to contain 6 elements (3 pairs) (1 full match pair and 2 sub-matches pair)"))
		}

		value := full[match[0]:match[1]]
		typeName := full[match[2]:match[3]]
		section := full[match[4]:match[5]]

		if traceEnabled {
			zlog.Debug("schema section", zap.String("value", value), zap.String("type", typeName), zap.String("section", section))
		}

		if typeName == "Query" {
			querySections = append(querySections, section)
		} else if typeName == "Subscription" {
			subscriptionSections = append(subscriptionSections, section)
		} else {
			panic(fmt.Errorf("invalid type name %s", typeName))
		}
	}

	// We need to do it reverse, the matches positions in text
	// are for the original file! By doing it reverse, we always
	// truncate from the end to the start ensuring truncation
	// position before the truncation `i` are still valid.
	for i := len(matches) - 1; i >= 0; i-- {
		full = truncate(full, matches[i][0], matches[i][1])
	}

	if len(querySections) > 0 {
		full = full + "\n" + `"""Query holds the top-level calls that return single responses of data. Call them with the 'query' keyword from your top-level GraphQL document"""` + "\n"
		full = full + fmt.Sprintf("type Query {\n%s\n}", strings.Join(querySections, "\n"))
	}

	if len(subscriptionSections) > 0 {
		full = full + "\n" + `"""Subscription holds the top-level calls that return streams of data. Call them with the 'subscription' keyword from your top-level GraphQL query"""` + "\n"
		full = full + fmt.Sprintf("type Subscription {\n%s\n}", strings.Join(subscriptionSections, "\n"))
	}

	return full
}

func truncate(in string, start, end int) (truncated string) {
	if start < 0 {
		panic(fmt.Errorf("start %d should be greater or equal to 0", start))
	}

	if end > len(in) {
		panic(fmt.Errorf("end %d should be less than %d", end, len(in)))
	}

	return in[0:start] + in[end:len(in)]
}

func parseSchema(name string, resolver interface{}, rawSchema string) (out *graphql.Schema, err error) {
	zlog.Info("parsing schema", zap.String("name", name), zap.String("resolver", fmt.Sprintf("%T", resolver)))

	out, err = graphql.ParseSchema(
		rawSchema,
		resolver,
		graphql.PrefixRootFunctions(),
		graphql.UseStringDescriptions(), graphql.UseFieldResolvers(),
		graphql.MaxDepth(24), // this is good for at least 6 levels of `inlineTraces`, fetching its data, etc..
		graphql.SubscribeResolverTimeout(10*time.Second),
		graphql.Tracer(&OpencensusTracer{}),
	)

	if err != nil && os.Getenv("TRACE") == "true" {
		// For debugging purposes, so that the full concatenated & merge schema is display to the developer
		fmt.Println()
		fmt.Printf("====================== %s Schema String ======================\n", name)
		printSchema(rawSchema)
		fmt.Println("====================================================================")
		fmt.Println()
	}

	return
}

func isCommonAssetName(name string) bool {
	return !strings.HasSuffix(name, "_alpha.graphql")
}

func isAlphaAssetName(name string) bool {
	return strings.HasSuffix(name, "_alpha.graphql")
}

func buildSchemaString(tag string, assetBytes [][]byte) *string {
	zlogger := zlog.With(zap.String("tag", tag))
	if len(assetBytes) == 0 {
		zlogger.Info("no asset data received, returning nil")
		return nil
	}

	zlogger.Info("combining asset data into GraphQL string schema", zap.Int("num_of_assets", len(assetBytes)))

	buf := bytes.Buffer{}
	for _, data := range assetBytes {
		buf.Write(data)

		// Add a newline if the file does not end in a newline.
		if len(data) > 0 && data[len(data)-1] != '\n' {
			buf.WriteByte('\n')
		}
	}

	schema := buf.String()

	return &schema
}

func printSchema(schema string) {
	lines := strings.Split(schema, "\n")
	maxDigitCount := digitCount(len(lines))

	for i, line := range lines {
		digitCount := digitCount(i + 1)
		padding := strings.Repeat(" ", maxDigitCount-digitCount+2)

		fmt.Printf("%d:%s%s\n", i, padding, line)
	}
}

func digitCount(i int) int {
	return int(math.Floor(math.Log10(float64(i) + 1)))
}

type withAlpha schemaID

func (s withAlpha) SchemaID() schemaID {
	return alpha
}

func (s withAlpha) String() string {
	return string(s)
}

func isSelectedSchema(candidate schemaID, selectors []SchemaSelector) bool {
	for _, selector := range selectors {
		if candidate == selector.SchemaID() {
			return true
		}
	}

	return false
}
