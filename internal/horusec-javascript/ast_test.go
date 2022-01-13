// Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package javascript_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ZupIT/horusec-engine/internal/ast"
	javascript "github.com/ZupIT/horusec-engine/internal/horusec-javascript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testcase struct {
	name string
	src  string
	file ast.File
}

func TestJavaScriptParseFile(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err, "Expected no error to get current working directory: %v", err)
	sources := filepath.Join(wd, "testdata", "source")
	expecteds := filepath.Join(wd, "testdata", "expected")

	err = filepath.Walk(sources, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		t.Run(info.Name(), func(t *testing.T) {

			src, err := os.ReadFile(path)
			require.NoError(t, err, "Expected no error to read source file: %v", err)

			f, err := javascript.ParseFile(info.Name(), src)
			require.NoError(t, err, "Expected no error to parse source file: %v", err)

			buf := bytes.NewBufferString("")

			ast.Fprint(buf, f)

			expectedAst, err := os.ReadFile(filepath.Join(expecteds, fmt.Sprintf("%s.out", info.Name())))
			require.NoError(t, err, "Expected no error to expected ast: %v", err)

			assert.Equal(t, string(expectedAst), buf.String())
		})

		return nil
	})

	assert.NoError(t, err, "Expected no error to walk on source code files: %v", err)
}
