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

package ast_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ZupIT/horusec-engine/internal/ast"
	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	testcases := []struct {
		n ast.Node
		e string
	}{
		{
			n: nil,
			e: "0  nil",
		},
		{
			n: &ast.Ident{
				Name: "foo",
			},
			e: `
0  *ast.Ident {
1  .  Name: "foo"
2  .  Position: ast.Position {}
3  }
			`,
		},
		{
			n: &ast.BasicLit{
				Kind:  "number",
				Value: "10",
			},
			e: `
0  *ast.BasicLit {
1  .  Position: ast.Position {}
2  .  Kind: "number"
3  .  Value: "10"
4  }
			`,
		},
		{
			n: &ast.FieldList{
				List: []*ast.Field{
					{
						Name: &ast.Ident{
							Name: "a",
						},
					},
					{
						Name: &ast.Ident{
							Name: "b",
						},
					},
				},
			},
			e: `
0  *ast.FieldList {
1  .  Position: ast.Position {}
2  .  List: []*ast.Field (len = 2) {
3  .  .  0: *ast.Field {
4  .  .  .  Position: ast.Position {}
5  .  .  .  Name: *ast.Ident {
6  .  .  .  .  Name: "a"
7  .  .  .  .  Position: ast.Position {}
8  .  .  .  }
9  .  .  }
10  .  .  1: *ast.Field {
11  .  .  .  Position: ast.Position {}
12  .  .  .  Name: *ast.Ident {
13  .  .  .  .  Name: "b"
14  .  .  .  .  Position: ast.Position {}
15  .  .  .  }
16  .  .  }
17  .  }
18  }
			`,
		},
		{
			n: &ast.FuncLit{
				Type: &ast.FuncType{
					Params:  &ast.FieldList{},
					Results: nil,
				},
				Body: nil,
			},
			e: `
0  *ast.FuncLit {
1  .  Position: ast.Position {}
2  .  Type: *ast.FuncType {
3  .  .  Position: ast.Position {}
4  .  .  Params: *ast.FieldList {
5  .  .  .  Position: ast.Position {}
6  .  .  }
7  .  }
8  }
			`,
		},
	}

	var buf bytes.Buffer
	for _, tt := range testcases {
		buf.Reset()
		assert.NoError(t, ast.Fprint(&buf, tt.n))
		assert.Equal(t, trim(tt.e), trim(buf.String()), "Expected:\n%s\n\nGot:\n%s", trim(tt.e), trim(buf.String()))
	}
}

// trim Split s into lines, trim whitespace from all lines, and return
// the concatenated non-empty lines.
//
// NOTE: This code was get from go/ast/print_test.go file.
func trim(s string) string {
	lines := strings.Split(s, "\n")
	i := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			lines[i] = line
			i++
		}
	}
	return strings.Join(lines[0:i], "\n")
}
