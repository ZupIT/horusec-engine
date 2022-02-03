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

// appendLocal create a new local Var on function context to a given ident and value.
func (fn *Function) appendLocal(ident *ast.Ident, value ast.Expr) {
	fn.Locals[ident.Name] = &Var{
		name:  ident.Name,
		Value: exprValue(value),
	}
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
		// Handle a, b = 1, 2
		if len(stmt.LHS) == len(stmt.RHS) {
			for idx := range stmt.LHS {
				lh, rh := stmt.LHS[idx], stmt.RHS[idx]
				// TODO(matheus): lh will always be an *ast.Ident?
				fn.appendLocal(lh.(*ast.Ident), rh)
			}

			break
		}
		// TODO(matheus): Handle cases like a, b = foo()
		panic("ir.builder.stmt: not implemented tuple assignments")
	default:
		panic(fmt.Sprintf("ir.builder.stmt: unhandled expression type: %T", stmt))
	}
}

// expr convert an expression e to a IR form.
//
func (b *builder) expr(fn *Function, e ast.Expr) {
	switch expr := e.(type) {
	case *ast.CallExpr:
		fn.emit(newCall(fn, expr))
	default:
		panic(fmt.Sprintf("ir.builder.expr: unhandled expression type: %T", expr))
	}
}

// stmtList emits to fn code for all statements in list.
func (b *builder) stmtList(fn *Function, list []ast.Stmt) {
	for _, s := range list {
		b.stmt(fn, s)
	}
}
