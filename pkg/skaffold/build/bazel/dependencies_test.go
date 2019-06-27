/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bazel

import (
	"context"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestGetDependencies(t *testing.T) {
	tests := []struct {
		description   string
		target        string
		files         map[string]string
		expectedQuery string
		output        string
		expected      []string
	}{
		{
			description:   "with workspace",
			target:        "target",
			files:         map[string]string{"WORKSPACE": "content is ignored"},
			expectedQuery: "bazel query kind('source file', deps('target')) union buildfiles('target') --noimplicit_deps --order_output=no",
			output:        "@ignored\n//external/ignored\n\n//:dep1\n//:dep2\n",
			expected:      []string{"dep1", "dep2", "WORKSPACE"},
		},
		{
			description:   "without workspace",
			target:        "target2",
			files:         map[string]string{},
			expectedQuery: "bazel query kind('source file', deps('target2')) union buildfiles('target2') --noimplicit_deps --order_output=no",
			output:        "@ignored\n//external/ignored\n\n//:dep3\n",
			expected:      []string{"dep3"},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.Override(&util.DefaultExecCommand, t.FakeRunOut(test.expectedQuery, test.output))
			tmpDir := t.NewTempDir().
				WriteFiles(test.files)

			deps, err := GetDependencies(context.Background(), tmpDir.Root(), &latest.BazelArtifact{
				BuildTarget: test.target,
			})

			t.CheckNoError(err)
			t.CheckDeepEqual(test.expected, deps)
		})
	}
}

func TestQuery(t *testing.T) {
	query := query("//:skaffold_example.tar")

	expectedQuery := `kind('source file', deps('//:skaffold_example.tar')) union buildfiles('//:skaffold_example.tar')`
	if query != expectedQuery {
		t.Errorf("Expected [%s]. Got [%s]", expectedQuery, query)
	}
}

func TestDepToPath(t *testing.T) {
	var tests = []struct {
		description string
		dep         string
		expected    string
	}{
		{
			description: "top level file",
			dep:         "//:dispatcher.go",
			expected:    "dispatcher.go",
		},
		{
			description: "vendored file",
			dep:         "//vendor/github.com/gorilla/mux:mux.go",
			expected:    "vendor/github.com/gorilla/mux/mux.go",
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			path := depToPath(test.dep)

			t.CheckDeepEqual(test.expected, path)
		})
	}
}
