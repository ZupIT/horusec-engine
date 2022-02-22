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

// nolint:funlen // We need a lot of lines and if's to convert an AST to IR.
package ir

import (
	"fmt"

	"github.com/ZupIT/horusec-engine/internal/ast"
)

// Build build all function members of file f.
//
// TODO(matheus): Decide how to deal with top level expressions on f.expresions.
func (f *File) Build() {
	for _, member := range f.Members {
		if fn, ok := member.(*Function); ok {
			fn.Build()
		}
	}
}

// Build the IR code for this function.
//
// nolint: stylecheck // Linter ask to change fn receiver to f to
// follow the same signature of File.Build, but this is not necessary.
func (fn *Function) Build() {
	var b builder

	var body *ast.BlockStmt

	switch s := fn.syntax.(type) {
	case *ast.FuncDecl:
		body = s.Body
	case *ast.FuncLit:
		body = s.Body
	default:
		panic(fmt.Sprintf("ir.Function.Build: invalid syntax node of function: %T", s))
	}

	b.buildFunction(fn, body)
}

// newBasicBlock adds to fn a new basic block and returns it.
// comment is an optional string for more readable debugging output.
//
// It does not automatically become the current block for subsequent calls to emit.
func (fn *Function) newBasicBlock(comment string) *BasicBlock {
	b := &BasicBlock{
		Index:   len(fn.Blocks),
		Comment: comment,
		Instrs:  make([]Instruction, 0),
	}

	fn.Blocks = append(fn.Blocks, b)

	return b
}

// emit add a new instruction on current basic block of function.
func (fn *Function) emit(instr Instruction) {
	fn.currentBlock.Instrs = append(fn.currentBlock.Instrs, instr)
}

// addNamedLocal creates a local variable, adds it to function fn and return it.
//
// Subsequent calls to fn.lookup(ident.Name) will return the same variable.
func (fn *Function) addNamedLocal(name *ast.Ident, value ast.Expr) Value {
	v := &Var{
		node:  node{name},
		name:  name.Name,
		Value: exprValue(fn, value),
	}
	fn.Locals[name.Name] = v
	fn.emit(v)
	return v
}

// addLocal create a new temporary variable, adds it to function fn and return it.
//
// The temporary name used is %tN where N is the current number of local variables
// on fn. The % prefix is added on variable name to avoid collisions.
func (fn *Function) addLocal(value Value, syntax ast.Node) Value {
	name := fmt.Sprintf("%%t%d", len(fn.Locals))
	v := &Var{
		node:  node{syntax},
		name:  name,
		Value: value,
	}
	fn.Locals[name] = v
	fn.emit(v)
	return v
}

// builder controls how a function is converted from AST to a IR.
//
// Its methods contain all the logic for AST-to-IR conversion.
type builder struct{}

// buildFunction builds IR code for the body of function fn.
func (b *builder) buildFunction(fn *Function, body *ast.BlockStmt) {
	fn.currentBlock = fn.newBasicBlock("entry")
	b.stmt(fn, body)
	fn.finishBody()
}

// finishBody finalizes the function after IR code generation of its body.
func (fn *Function) finishBody() {
	fn.currentBlock = nil
}

// stmt convert a statement s to a IR form.
//
// nolint:gocyclo // Its better centralize all stmt to IR conversion on a single function.
func (b *builder) stmt(fn *Function, s ast.Stmt) {
	switch stmt := s.(type) {
	case *ast.BlockStmt:
		b.stmtList(fn, stmt.List)
	case *ast.ExprStmt:
		b.expr(fn, stmt.Expr)
	case *ast.AssignStmt:
		b.assignStmt(fn, stmt.LHS, stmt.RHS)
	case *ast.ReturnStmt:
		results := make([]Value, 0, len(stmt.Results))

		for _, res := range stmt.Results {
			results = append(results, exprValue(fn, res))
		}

		fn.emit(&Return{Results: results, node: node{stmt}})
		fn.currentBlock = fn.newBasicBlock("unreachable")
	case *ast.BadNode:
		// Do nothing with bad nodes.
	default:
		panic(fmt.Sprintf("ir.builder.stmt: unhandled expression type: %T", stmt))
	}
}

// expr convert an expression e to a IR form.
func (b *builder) expr(fn *Function, e ast.Expr) {
	switch expr := e.(type) {
	case *ast.CallExpr:
		fn.emit(callExpr(fn, expr))
	default:
		panic(fmt.Sprintf("ir.builder.expr: unhandled expression type: %T", expr))
	}
}

// assignStmt emits code to fn for a parallel assignment of rhss to lhss.
func (b *builder) assignStmt(fn *Function, lhss, rhss []ast.Expr) {
	if len(lhss) == len(rhss) {
		// Simple assignment:      x     = f()
		// or Parallel assignment: x, y  = f(), g()
		for idx := range lhss {
			b.assign(fn, lhss[idx], rhss[idx])
		}

		return
	}
	// TODO(matheus): Handle cases like a, b = foo()
	panic("ir.builder.assignStmt: not implemented tuple assignments")
}

// assign emits to fn code to initialize the lhs with the value
// of expression rhs.
func (b *builder) assign(fn *Function, lhs, rhs ast.Expr) {
	switch lhs := lhs.(type) {
	case *ast.Ident:
		if closure, isFuncLit := rhs.(*ast.FuncLit); isFuncLit {
			fn.emit(funcLit(fn, lhs.Name, closure))

			return
		}
		fn.addNamedLocal(lhs, rhs)
	default:
		panic(fmt.Sprintf("ir.builder.assingStmt: not handled lhs assignment type: %T", lhs))
	}
}

// stmtList emits to fn code for all statements in list.
func (b *builder) stmtList(fn *Function, list []ast.Stmt) {
	for _, s := range list {
		b.stmt(fn, s)
	}
}
