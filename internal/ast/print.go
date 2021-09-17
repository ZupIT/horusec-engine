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

package ast

import (
	"go/ast"
	"io"
	"os"
)

// Print prints n to standard output, skipping nil fields.
// Print(n) is the same as Fprint(os.Stdout, n).
func Print(n Node) error {
	return Fprint(os.Stdout, n)
}

// Fprint prints n to w output skipping nil fields.
func Fprint(w io.Writer, n Node) error {
	// The ast.Fprint implementation from go/ast package is a
	// generic implementation to print arbitrary data structures.
	// Since the data structure used to print here is an AST too,
	// we can reuse the Go std library implementation.
	return ast.Fprint(w, nil, n, ast.NotNilFilter)
}
