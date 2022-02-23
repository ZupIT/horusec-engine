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
		case *ast.ValueDecl:
			for _, g := range valueDecl(decl) {
				file.Members[g.Name()] = g
			}
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
func (f *File) NewFunction(decl *ast.FuncDecl) *Function {
	fn := &Function{
		name:      decl.Name.Name,
		syntax:    decl,
		File:      f,
		Blocks:    make([]*BasicBlock, 0),
		Locals:    make(map[string]*Var),
		AnonFuncs: make([]*Function, 0),
		parent:    nil,
	}

	return fn
}

// exprValue lowers a single-result expression e to IR form and return the Value defined by the expression.
//
// nolint: gocyclo // Some checks is needed.
func exprValue(parent *Function, e ast.Expr) Value {
	switch expr := e.(type) {
	case *ast.BasicLit:
		return &Const{
			node:  node{e},
			Value: expr.Value,
		}
	case *ast.Ident:
		return &Var{
			node:  node{e},
			name:  expr.Name,
			Value: nil,
		}
	case *ast.CallExpr:
		return callExpr(parent, expr)
	case *ast.BinaryExpr:
		return binaryExpr(parent, expr)
	case *ast.FuncLit:
		// Create an anonymous function using the parent function name
		// and the current the total of anonymouns functions as a name.
		return funcLit(parent, fmt.Sprintf("%s$%d", parent.Name(), len(parent.AnonFuncs)+1), expr)
	case *ast.TemplateExpr:
		values := make([]Value, 0, len(expr.Subs))
		for _, v := range expr.Subs {
			values = append(values, exprValue(parent, v))
		}
		return &Template{
			node:  node{expr},
			Value: expr.Value,
			Subs:  values,
		}
	default:
		unsupportedNode(expr)
		return nil
	}
}

// funcLit crate a new Closure to a given AST based function literal.
func funcLit(parent *Function, name string, syntax *ast.FuncLit) *Closure {
	fn := &Function{
		name:   name,
		syntax: syntax,
		File:   parent.File,
		parent: parent,
		Blocks: make([]*BasicBlock, 0),
		Locals: make(map[string]*Var),
	}
	fn.Build()

	parent.AnonFuncs = append(parent.AnonFuncs, fn)

	return &Closure{fn}
}

// binaryExpr create a new IR BinOp from the given binary expression expr.
//
// The operands of binary expression is resolved recursively.
func binaryExpr(parent *Function, expr *ast.BinaryExpr) *BinOp {
	return &BinOp{
		node:  node{expr},
		Op:    expr.Op,
		Left:  exprValue(parent, expr.Left),
		Right: exprValue(parent, expr.Right),
	}
}

// callExpr create new Call to a given ast.CallExpr
//
// If CallExpr arguments use a variable declared inside parent function
// call arguments will point to to this declared variable.
//
// nolint:gocyclo // Some checks is needed here.
func callExpr(parent *Function, call *ast.CallExpr) *Call {
	args := make([]Value, 0, len(call.Args))

	for _, arg := range call.Args {
		var v Value
		switch arg := arg.(type) {
		case *ast.Ident:
			// If identifier used on function call is declared inside the parent function
			// we use this declared variable as argument to function call, otherwise if
			// just parse the expression value.
			if local := parent.lookup(arg.Name); local != nil {
				v = local
			} else {
				v = exprValue(parent, arg)
			}
		case *ast.CallExpr:
			// If the argument value is a call expression we first create a temp variable
			// assigin your result to this function call, and them set this temp variable
			// as an argument to function call.
			v = parent.addLocal(callExpr(parent, arg), arg)
		default:
			v = exprValue(parent, arg)
		}
		args = append(args, v)
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
			unsupportedNode(call.Expr)
			break
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
		unsupportedNode(call)
	}

	return &Call{
		node: node{
			syntax: call,
		},
		Parent:   parent,
		Function: fn,
		Args:     args,
	}
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
