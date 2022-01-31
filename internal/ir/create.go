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

// NewFunction create a new Function to a given function declaration.
//
// The real work of building the IR form for a function is not done
// until a call to Function.Build().
func NewFunction(decl *ast.FuncDecl) *Function {
	var (
		params  []*Parameter
		results []*Parameter
		fn      = &Function{
			Name:   decl.Name.Name,
			syntax: decl,
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
			Name:   expr.Name,
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
// nolint:gocritic // More switch cases will be added.
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

	var fn Function

	switch call := call.Fun.(type) {
	case *ast.Ident:
		// TODO(matheus): Check if this function is declared inside the module that parent belongs
		// if its presents use this function function pointer on Call structure.
		//
		// Note: We still don't have a module representation to fix this.
		fn.Name = call.Name
	}

	return &Call{
		Parent:   parent,
		Function: &fn,
		Args:     args,
	}
}
