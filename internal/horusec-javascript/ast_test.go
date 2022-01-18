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
	"flag"
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

func TestMain(m *testing.M) {
	rewrite := flag.Bool("rewrite", false, "Rewrite expected AST output files")
	flag.Parse()

	if *rewrite {
		fmt.Println("rewriting expected ASTs")
		if err := rewriteExpectedASTs(); err != nil {
			fmt.Fprintf(os.Stderr, "Error to rewrite ASTs: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

type testcase struct {
	name string
	src  string
	file ast.File
}

func TestJavaScriptParseFileFillAllPositions(t *testing.T) {
	src := []byte(`
class Foo { f(a, b) { return a + b }}

function f1(s) { console.log(s) } 

const f2 (a, b) => { return a / b }
	`)

	f, err := javascript.ParseFile("", src)
	require.NoError(t, err, "Expected no error to parse source file: %v", err)

	notExpectedPos := ast.Pos{
		Byte:   0,
		Row:    0,
		Column: 0,
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		start := f.Start()
		end := f.End()

		assert.NotEqual(t, notExpectedPos, start, "Expected not empty start position from node %T: %s", n, start)
		assert.NotEqual(t, notExpectedPos, end, "Expected not empty end position from node %T: %s", n, end)

		return true
	})
}

func TestJavaScriptParseFile(t *testing.T) {
	sources := sourceFilesPath()
	err := filepath.Walk(sources, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		t.Run(info.Name(), func(t *testing.T) {
			f, err := astFromFile(info, path)
			require.NoError(t, err, "Expected no error to parse source file: %v", err)

			buf := bytes.NewBufferString("")

			ast.Fprint(buf, f)

			expectedAst, err := os.ReadFile(expectedFilePath(info))
			require.NoError(t, err, "Expected no error to expected ast: %v", err)

			assert.Equal(t, string(expectedAst), buf.String())
		})

		return nil
	})

	assert.NoError(t, err, "Expected no error to walk on source code files: %v", err)
}

func astFromFile(info fs.FileInfo, path string) (*ast.File, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return javascript.ParseFile(info.Name(), src)
}

// rewriteExpectedASTs rewrite all expected AST files on testdata/expected to all source
// all inside testdata/sources.
func rewriteExpectedASTs() error {
	err := filepath.Walk(sourceFilesPath(), func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		f, err := astFromFile(info, path)
		if err != nil {
			return err
		}

		expectedFile, err := os.OpenFile(expectedFilePath(info), os.O_WRONLY, fs.ModePerm)
		if err != nil {
			return err
		}
		defer expectedFile.Close()

		return ast.Fprint(expectedFile, f)
	})

	return err
}

// sourceFilesPath return the absoulute path that contains all source files.
func sourceFilesPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "testdata", "source")
}

// expectedFilePath return the absolute path of expected AST to a given source file info.
func expectedFilePath(info fs.FileInfo) string {
	return filepath.Join(expectedFilesPath(), fmt.Sprintf("%s.out", info.Name()))
}

// expectedFilesPath return the absoulute path that contains all expected AST files.
func expectedFilesPath() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "testdata", "expected")
}
