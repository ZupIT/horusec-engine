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
// TODO(matheus): Decide how to deal with top level expressions on f.expressions.
func (f *File) Build() {
	for _, member := range f.Members {
		if fn, ok := member.(*Function); ok {
			fn.Build()
		}
	}
}

// Build the IR code for this function.
func (fn *Function) Build() {
	var b builder

	var (
		body     *ast.BlockStmt
		funcType *ast.FuncType
	)

	switch s := fn.syntax.(type) {
	case *ast.FuncDecl:
		body = s.Body
		funcType = s.Type
	case *ast.FuncLit:
		body = s.Body
		funcType = s.Type
	default:
		panic(fmt.Sprintf("ir.Function.Build: invalid syntax node of function: %T", s))
	}

	fn.Signature = b.buildFuncSignature(fn, funcType)
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
//
// If the instruction defines a Value, it is returned.
func (fn *Function) emit(instr Instruction) Value {
	fn.currentBlock.Instrs = append(fn.currentBlock.Instrs, instr)
	v, _ := instr.(Value)
	return v
}

// addNamedLocal creates a local variable, adds it to function fn and return it.
//
// Subsequent calls to fn.lookup(name) will return the same variable.
func (fn *Function) addNamedLocal(name string, value Value, syntax ast.Node) Value {
	v := &Var{
		node:  node{syntax},
		name:  name,
		Value: value,
	}
	fn.Locals[name] = v
	fn.emit(v)
	return v
}

// addLocal create a new temporary variable, adds it to function fn and return it.
//
// The temporary name used is %tN where N is the current number of local variables
// on fn. The % prefix is added on variable name to avoid collisions.
func (fn *Function) addLocal(value Value, syntax ast.Node) Value {
	name := fmt.Sprintf("%%t%d", len(fn.Locals))
	return fn.addNamedLocal(name, value, syntax)
}

// builder controls how a function is converted from AST to an IR.
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

// stmt convert a statement s to an IR form.
//
// nolint:gocyclo // Its better centralize all stmt to IR conversion on a single function.
func (b *builder) stmt(fn *Function, s ast.Stmt) {
	switch stmt := s.(type) {
	case *ast.BlockStmt:
		b.stmtList(fn, stmt.List)
	case *ast.ExprStmt:
		b.expr(fn, stmt.Expr, true /*expand*/)
	case *ast.AssignStmt:
		b.assignStmt(fn, stmt.LHS, stmt.RHS, stmt)
	case *ast.ReturnStmt:
		results := make([]Value, 0, len(stmt.Results))

		for _, res := range stmt.Results {
			results = append(results, b.expr(fn, res, true /*expand*/))
		}

		fn.emit(&Return{Results: results, node: node{stmt}})
		fn.currentBlock = fn.newBasicBlock("unreachable")
	case *ast.BadNode:
	// Do nothing with bad nodes.
	case *ast.IfStmt:
		b.ifStmt(fn, stmt)
	default:
		unsupportedNode(stmt)
	}
}

// expr lowers a single-result expression e to IR form and return the Value defined by the expression.
//
// The expand parameter informs if expr should expand the expression e to a temporary variable. If
// expand is true and the expression e lowers to an IR Value that is also an Instruction, a temporary
// variable will be created using the Value created from e as the variable value, and the returned
// Value will be the created temporary variable. Otherwise the IR Value created from e is returned.
//
// Note that Var node is an exception to the rule informed above, because it is already a variable.
func (b *builder) expr(fn *Function, e ast.Expr, expand bool) Value {
	var v Value

	switch expr := e.(type) {

	// Value's that are *not* Instruction's (Var is an exception, see the doc above)
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
	case *ast.TemplateExpr:
		values := make([]Value, 0, len(expr.Subs))
		for _, v := range expr.Subs {
			values = append(values, b.expr(fn, v, expand))
		}
		return &Template{
			node:  node{expr},
			Value: expr.Value,
			Subs:  values,
		}

	// Value's that are also Instruction's.
	case *ast.FuncLit:
		// Create an anonymous function using the parent function name
		// and the current the total of anonymous functions as a name.
		v = b.funcLit(fn, fmt.Sprintf("%s$%d", fn.Name(), len(fn.AnonFuncs)+1), expr)
	case *ast.CallExpr:
		v = b.callExpr(fn, expr)
	case *ast.BinaryExpr:
		v = b.binaryExpr(fn, expr)
	default:
		unsupportedNode(expr)
		return nil
	}

	if expand {
		return fn.addLocal(v, e)
	}

	return v
}

// assignStmt emits code to fn for a parallel assignment of rhss to lhss.
func (b *builder) assignStmt(fn *Function, lhss, rhss []ast.Expr, syntax *ast.AssignStmt) {
	if len(lhss) == len(rhss) {
		// Simple assignment:      x     = f()
		// or Parallel assignment: x, y  = f(), g()
		for idx := range lhss {
			b.assign(fn, lhss[idx], rhss[idx], syntax)
		}

		return
	}
	// TODO(matheus): Handle cases like a, b = foo()
	if debugIsEnable() {
		panic("ir.builder.assignStmt: not implemented tuple assignments")
	}
}

// assign emits to fn code to initialize the lhs with the value
// of expression rhs.
func (b *builder) assign(fn *Function, lhs, rhs ast.Expr, syntax *ast.AssignStmt) {
	switch lhs := lhs.(type) {
	case *ast.Ident:
		if closure, isFuncLit := rhs.(*ast.FuncLit); isFuncLit {
			fn.emit(b.funcLit(fn, lhs.Name, closure))

			return
		}
		fn.addNamedLocal(lhs.Name, b.expr(fn, rhs, false /*expand*/), syntax)
	default:
		unsupportedNode(lhs)
	}
}

// stmtList emits to fn code for all statements in list.
func (b *builder) stmtList(fn *Function, list []ast.Stmt) {
	for _, s := range list {
		b.stmt(fn, s)
	}
}

// funcLit crate a new Closure to a given AST based function literal.
func (b *builder) funcLit(parent *Function, name string, syntax *ast.FuncLit) *Closure {
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
func (b *builder) binaryExpr(parent *Function, expr *ast.BinaryExpr) *BinOp {
	return &BinOp{
		node:  node{expr},
		Op:    expr.Op,
		Left:  b.expr(parent, expr.Left, true /* expand */),
		Right: b.expr(parent, expr.Right, true /* expand */),
	}
}

// callExpr create new Call to a given ast.CallExpr
//
// If CallExpr arguments use a variable declared inside parent function
// call arguments will point to this declared variable.
//
// nolint:gocyclo // Some checks is needed here.
func (b *builder) callExpr(parent *Function, call *ast.CallExpr) *Call {
	args := make([]Value, 0, len(call.Args))

	for _, arg := range call.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			// If identifier used on function call is declared inside the parent function
			// we use this declared variable as argument to function call, otherwise if
			// just parse the expression value.
			if local := parent.lookup(ident.Name); local != nil {
				args = append(args, local)
				continue
			}
		}
		args = append(args, b.expr(parent, arg, true /* expand */))
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
			// TODO(matheus): Add support to call expressions using selector expressions.
			// e.g: a.b.c()
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

// buildFuncSignature build function signature of fn to a given function type.
//
// NOTE: buildFuncSignature don't set fn.Signature field, it just build the IR
// representation of parameters and results.
func (b *builder) buildFuncSignature(fn *Function, funcType *ast.FuncType) *Signature {
	var (
		params  []*Parameter
		results []*Parameter
	)

	if funcType.Params != nil {
		params = make([]*Parameter, 0, len(funcType.Params.List))
		for _, p := range funcType.Params.List {
			params = append(params, b.buildFuncParameter(fn, p.Name))
		}
	}

	if funcType.Results != nil {
		results = make([]*Parameter, 0, len(funcType.Results.List))
		for _, p := range funcType.Results.List {
			results = append(results, b.buildFuncParameter(fn, p.Name))
		}
	}

	return &Signature{params, results}
}

// buildFuncParameter build a new function parameter to a given expression.
func (b *builder) buildFuncParameter(fn *Function, expr ast.Expr) *Parameter {
	switch expr := expr.(type) {
	case *ast.Ident:
		return &Parameter{
			parent: fn,
			name:   expr.Name,
			Value:  nil,
		}
	case *ast.ObjectExpr:
		var v Value
		if len(expr.Elts) > 0 {
			// Since default parameter values can not have more than
			// one value, we check if the value really exists and use
			// it to create the parameter value.
			v = b.expr(fn, expr.Elts[0], false /*expand*/)
		}
		return &Parameter{
			parent: fn,
			name:   expr.Name.Name,
			Value:  v,
		}
	default:
		unsupportedNode(expr)
		return nil
	}
}

// ifStmt
func (b *builder) ifStmt(fn *Function, stmt *ast.IfStmt) {
	b.expr(fn, stmt.Cond, true)
	b.stmt(fn, stmt.Body)

	if stmt.Else != nil {
		b.stmt(fn, stmt.Else)
	}
}
