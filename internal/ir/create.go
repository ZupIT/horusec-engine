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
	// TODO(matheus): Decide how to support expressions that is not
	// inside a declaration.
	if len(f.Exprs) > 0 {
		panic("ir.NewFile: AST expressions it not handled yet")
	}

	file := &File{
		Members:  make(map[string]Member),
		imported: make(map[string]*ExternalMember),
		name:     f.Name.Name,
	}

	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			fn := file.NewFunction(decl)
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
			file.Members[importt.Name()] = importt
			file.imported[importt.Name()] = importt
		default:
			panic(fmt.Sprintf("ir.NewFile: unhadled declaration type: %T", decl))
		}
	}

	return file
}

// NewFunction create a new Function to a given function declaration.
//
// The real work of building the IR form for a function is not done
// until a call to Function.Build().
func (f *File) NewFunction(decl *ast.FuncDecl) *Function {
	var (
		params  []*Parameter
		results []*Parameter
		fn      = &Function{
			name:   decl.Name.Name,
			syntax: decl,
			File:   f,
			Blocks: make([]*BasicBlock, 0),
			Locals: make(map[string]*Var),
		}
	)

	if decl.Type.Params != nil {
		params = make([]*Parameter, 0, len(decl.Type.Params.List))
		for _, p := range decl.Type.Params.List {
			params = append(params, newParameter(fn, p.Name))
		}
	}

	if decl.Type.Results != nil {
		results = make([]*Parameter, 0, len(decl.Type.Results.List))
		for _, p := range decl.Type.Results.List {
			results = append(results, newParameter(fn, p.Name))
		}
	}

	fn.Signature = &Signature{params, results}

	return fn
}

// newParameter return a new Parameter to a given expression.
func newParameter(fn *Function, expr ast.Expr) *Parameter {
	switch expr := expr.(type) {
	case *ast.Ident:
		return &Parameter{
			parent: fn,
			name:   expr.Name,
			Value:  nil,
		}
	default:
		panic(fmt.Sprintf("ir.newParameter: unhandled expression type: %T", expr))
	}
}

// exprValue lowers a single-result expression e to IR form and return the Value defined by the expression.
func exprValue(e ast.Expr) Value {
	switch expr := e.(type) {
	case *ast.BasicLit:
		return &Const{expr.Value}
	case *ast.Ident:
		return &Var{
			name:  expr.Name,
			Value: nil,
		}
	default:
		panic(fmt.Sprintf("ir.newValue: unhandled expression type: %T", expr))
	}
}

// newCall create new Call to a given ast.CallExpr
//
// If CallExpr arguments use a variable declared inside parent function
// call arguments will point to to this declared variable.
//
// nolint:gocyclo // Some checks is needed here.
func newCall(parent *Function, call *ast.CallExpr) *Call {
	args := make([]Value, 0, len(call.Args))

	for _, arg := range call.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			// If identifier used on function call is declared inside the parent function
			// we use this declared variable as argument to function call.
			if local, exists := parent.Locals[ident.Name]; exists {
				args = append(args, local)

				continue
			}
		}
		args = append(args, exprValue(arg))
	}

	fn := new(Function)

	switch call := call.Fun.(type) {
	case *ast.Ident:
		// TODO(matheus): This will not work if function is defined inside parent.
		if f := parent.File.Func(call.Name); f != nil {
			fn = f

			break
		}
		fn.name = call.Name
	case *ast.SelectorExpr:
		expr, ok := call.Expr.(*ast.Ident)
		if !ok {
			panic(fmt.Sprintf("ir.newCall: unhandled type of expression field from SelectorExpr: %T", call.Expr))
		}

		var ident string

		// Expr.Name could be an alias imported name, so need to check if this
		// identifier is imported so we use your real name. Otherwise we just
		// use the expression identifier name.
		if importt := parent.File.ImportedPackage(expr.Name); importt != nil {
			ident = importt.name
		} else {
			ident = expr.Name
		}

		fn.name = fmt.Sprintf("%s.%s", ident, call.Sel.Name)
	default:
		panic(fmt.Sprintf("ir.newCall: unhandled type of call function: %T", call))
	}

	return &Call{
		Parent:   parent,
		Function: fn,
		Args:     args,
	}
}

func identNameIfNotNil(i *ast.Ident) string {
	if i != nil {
		return i.Name
	}
	return ""
}
