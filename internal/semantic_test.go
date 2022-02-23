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

package semantic_test

// This file provide tests to assert that to a given source file, we correctly generate the generic
// AST and the IR representation of this file.
//
// TestParseAST is responsable to assert that source files is correctly parsed to the generic AST.
// TestParseIR is responsable to assert that source files is correctly parsed to IR.
//
// To provide theses tests in a generic way, both of these tests use the files created inside testdata
// directory as a source input and expected output of AST and IR. This directory is separated into two
// sub directories:
// - source: contains source files of languages that should be tested.
// - expected: contains the expected files representing the AST and IR of the files within source.
//
// These directories is divided by languages, e.g testdata/source/javascript and testdata/expected/javascript.
//
// The expected directory contains two sub-directories; ast and ir, which contains the expected files for
// each parsing process.
//
// The expected files **should** contains the same name inside source directory using the .out extesion,
// e.g testdata/source/javascript/main.js > testdata/expected/javascript/ast/main.js.out, so to add a new
// testcase for an AST parsing it is only necessary to create a file in source/<language> and a file with
// the expected AST/IR in expected/<language>/ast or expected/<language>/ir.
//
// Since writing the expected AST and IR files would be a borring task, these tests expose a -rewrite flag
// that can be used to create new testcases and update the existing ones. When this flag is passed, before
// executing the testcase the expected file will be rewrited with the value that the parsing process return.
// So to add a new testcase just create the source file and them run the tests using the -rewrite flag:
// go test -v ./internal/ -args -rewrite

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ZupIT/horusec-engine/internal/ast"
	javascript "github.com/ZupIT/horusec-engine/internal/horusec-javascript"
	"github.com/ZupIT/horusec-engine/internal/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// rootExpectedPath is the absoulute path of testdata/source path.
	rootSourcePath string

	// rootExpectedPath is the absolute path of testdata/expected path.
	rootExpectedPath string

	// rewrite is the rewrite flag used to rewrite expected files.
	rewrite = flag.Bool("rewrite", false, "Rewrite expected output files")
)

func TestMain(m *testing.M) {
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current working directory: %v", err)
		os.Exit(1)
	}

	rootSourcePath = filepath.Join(wd, "testdata", "source")
	rootExpectedPath = filepath.Join(wd, "testdata", "expected")

	os.Exit(m.Run())
}

// testcase contains the information needed to run the tests of
// parsing a source code to AST or IR representation.
type testcase struct {
	name         string                                           // Name of testcase, commonly the language name.
	parser       func(name string, src []byte) (*ast.File, error) // The parser function that create the ast.File.
	sourcePath   string                                           // Path that contains the source files.
	expectedPath string                                           // Path that contains the expected files.
}

func TestParseAST(t *testing.T) {
	testcases := []testcase{
		{
			name:         "Javascript",
			parser:       javascript.ParseFile,
			sourcePath:   filepath.Join(rootSourcePath, "javascript"),
			expectedPath: filepath.Join(rootExpectedPath, "javascript", "ast"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			testParseFiles(t, tt.sourcePath, tt.expectedPath, func(t *testing.T, name, path string) string {
				src, err := os.ReadFile(path)
				require.NoError(t, err, "Expected no error to read file: %v", err)

				f, err := tt.parser(name, src)
				require.NoError(t, err, "Expected no error to read file: %v", err)

				buf := bytes.NewBufferString("")

				ast.Fprint(buf, f)

				return buf.String()
			})
		})
	}
}

func TestParseIR(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping IR parsing tests")
	}
	testcases := []testcase{
		{
			name:         "Javascript",
			parser:       javascript.ParseFile,
			sourcePath:   filepath.Join(rootSourcePath, "javascript"),
			expectedPath: filepath.Join(rootExpectedPath, "javascript", "ir"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			testParseFiles(t, tt.sourcePath, tt.expectedPath, func(t *testing.T, name, path string) string {
				// Since mostly cases of AST to IR conversion is not fully implemented yet
				// the testcase will panic, so this recover prevents this and causes the
				// test to fail in a "nice" way.
				defer func() {
					if err := recover(); err != nil {
						t.Errorf("Panic on %s: %v", name, err)
					}
				}()
				buf := bytes.NewBufferString("")

				src, err := os.ReadFile(path)
				require.NoError(t, err, "Expected no error to read file: %v", err)

				ast, err := tt.parser(name, src)
				require.NoError(t, err, "Expected no error to read file: %v", err)

				irFile := ir.NewFile(ast)
				irFile.Build()

				irFile.WriteTo(buf)

				// Add a separation between file the IR representation and function body.
				buf.WriteString("\n\n============================\n\n")

				// Sort members to avoid out of order members break the test.
				memberNames := make([]string, 0, len(irFile.Members))
				for name := range irFile.Members {
					memberNames = append(memberNames, name)
				}
				sort.Strings(memberNames)

				for _, name := range memberNames {
					if fn, ok := irFile.Members[name].(*ir.Function); ok {
						fn.WriteTo(buf)
					}
				}

				return buf.String()
			})
		})
	}
}

// testParseFiles walk on source path comparing the return value from parse function with the expected
// value contained on expected path.
//
// The parse function will receive the name and the path of the source file that is being walked and
// should return the representation of this file that will be used to compare with the expected file.
//
// Read the documentation on the header of this file to understand the source and expected paths.
func testParseFiles(t *testing.T, source, expected string, parse func(t *testing.T, name string, path string) string) {
	err := filepath.Walk(source, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		t.Run(info.Name(), func(t *testing.T) {
			expectedFilePath := filepath.Join(expected, fmt.Sprintf("%s.out", info.Name()))

			output := parse(t, info.Name(), path)
			if *rewrite {
				rewriteExpectedFile(t, output, expectedFilePath)
			}

			expected, err := os.ReadFile(expectedFilePath)
			require.NoError(t, err, "Expected no error to read expected file: %v", err)

			assert.Equal(t, string(expected), output)
		})

		return nil
	})

	require.NoError(t, err, "Expected no error to walk on source path: %v", err)
}

func rewriteExpectedFile(t *testing.T, data, file string) {
	t.Logf("Rewriting expected file %s\n", file)
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, fs.ModePerm)
	require.NoError(t, err, "Expected no error to open file to rewrite: %v", err)

	err = f.Truncate(0)
	require.NoError(t, err, "Expected no error to truncate file: %v", err)

	_, err = f.Seek(0, io.SeekStart)
	require.NoError(t, err, "Expected no error to seek file to start: %v", err)

	_, err = f.WriteString(data)
	require.NoError(t, err, "Expected no error to write on file: %v", err)
}
