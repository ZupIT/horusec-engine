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

// nolint:funlen // We need a lot of lines and if to convert an AST to IR.
package ir

import (
	"fmt"

	"github.com/ZupIT/horusec-engine/internal/ast"
)

// NewFile create a new File to a given ast.File.
//
// The real work of building the IR form for a file is not done
// untila call to File.Build().
//
// NewFile only map function declarations and imports on retuned File.
//
// nolint:gocyclo // Some checks is needed here.
func NewFile(f *ast.File) *File {
	file := &File{
		Members:     make(map[string]Member),
		imported:    make(map[string]*ExternalMember),
		name:        f.Name.Name,
		expressions: f.Exprs,
		syntax:      f,
	}

	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			fn := file.NewFunction(decl.Name.Name, decl)
			if _, exists := file.Members[fn.Name()]; exists {
				panic(fmt.Sprintf("ir.NewFile: already existed function member: %s", fn.Name()))
			}
			file.Members[fn.Name()] = fn
		case *ast.ImportDecl:
			importt := &ExternalMember{
				name:  identNameIfNotNil(decl.Name),
				Path:  decl.Path.Name,
				Alias: identNameIfNotNil(decl.Alias),
			}
			file.Members[importt.ImportName()] = importt
			file.imported[importt.ImportName()] = importt
		case *ast.ValueDecl:
			for _, g := range valueDecl(decl) {
				file.Members[g.Name()] = g
			}
		case *ast.ClassDecl:
			s := &Struct{
				node: node{decl},
				name: decl.Name.Name,
			}

			for _, v := range decl.Body.List {
				switch d := v.(type) {
				case *ast.FuncDecl:
					s.Methods = append(s.Methods, file.NewFunction(d.Name.Name, d))
				case *ast.ValueDecl:
					values := valueDecl(d)
					for _, value := range values {
						s.Fields = append(s.Fields, value)
					}
				default:
					unsupportedNode(decl)
				}
			}

			file.Members[s.name] = s
		default:
			unsupportedNode(decl)
		}
	}

	return file
}

// NewFunction create a new Function to a given function declaration.
//
// The real work of building the IR form for a function is not done
// until a call to Function.Build().
func (f *File) NewFunction(name string, syntax ast.Node) *Function {
	fn := &Function{
		name:      name,
		syntax:    syntax,
		File:      f,
		Blocks:    make([]*BasicBlock, 0),
		Locals:    make([]*Var, 0),
		AnonFuncs: make([]*Function, 0),
		Signature: new(Signature),
		phis:      make(map[string]*Phi),
		nLocals:   0,
		parent:    nil,
	}

	return fn
}

// valueDecl create global variable declarations to a given value declaration.
//
// A new global declaration will be returned for each decl.Name and decl.Value.
//
// nolint: gocyclo // Some checks is needed.
func valueDecl(decl *ast.ValueDecl) []*Global {
	if len(decl.Names) < len(decl.Values) {
		panic("ir.create.newGlobals: global declaration values with more values than names")
	}

	globals := make([]*Global, 0, len(decl.Names))

	appendGlobal := func(ident *ast.Ident, value ast.Expr) {
		globals = append(globals, &Global{
			node:  node{decl},
			name:  ident.Name,
			Value: value,
		})
	}

	// Handle a, b = 1, 2
	if len(decl.Names) == len(decl.Values) {
		for idx := range decl.Names {
			appendGlobal(decl.Names[idx], decl.Values[idx])
		}
	} else {
		var value ast.Expr

		// Global variables can be declared without a initial value.
		if len(decl.Values) > 0 {
			value = decl.Values[0]
		}

		for _, name := range decl.Names {
			appendGlobal(name, value)
		}
	}

	return globals
}

func identNameIfNotNil(i *ast.Ident) string {
	if i != nil {
		return i.Name
	}
	return ""
}
