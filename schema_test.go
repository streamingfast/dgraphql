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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestSchema_MergeSchemas(t *testing.T) {
	graphqlResource := func(name string) string {
		document, err := ioutil.ReadFile("testdata/" + name)
		require.NoError(t, err)

		return string(document)
	}

	tests := []struct {
		name   string
		common string
		alpha  string
	}{
		{
			name:   "simple_queries",
			common: graphqlResource("simple_queries_common.graphql"),
			alpha:  graphqlResource("simple_queries_alpha.graphql"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if os.Getenv("TRACE") != "" {
				fmt.Println("Common Schema", test.common)
				fmt.Println("Alpha Schema", test.alpha)
			}

			goldenFile := fmt.Sprintf("testdata/%s.g.graphql", test.name)
			actual := MergeSchemas(test.alpha, test.common)
			if os.Getenv("GOLDEN_UPDATE") != "" {
				err := ioutil.WriteFile(goldenFile, []byte(actual), os.ModePerm)
				require.NoError(t, err)
			}

			expected, err := ioutil.ReadFile(goldenFile)
			require.NoError(t, err)

			assert.Equal(t, actual, string(expected))
		})
	}
}
