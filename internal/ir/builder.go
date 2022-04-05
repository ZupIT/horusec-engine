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

// Build builds all function members of file f.
func (f *File) Build() {
	for _, member := range f.Members {
		switch m := member.(type) {
		case *Function:
			m.Build()
		case *Struct:
			m.Build()
		}
	}

	if len(f.expressions) > 0 {
		f.buildExpressions()
	}
}

// buildExpressions creates a new function with all expressions parsed in it
//
// To build the expressions a temporary function will be created, to keep the same name pattern with the temporary
// variables the name will be in the same pattern (%fn0). As we are creating a function without a function AST.Node,
// it is not possible to use the function.Build function as usual, so we just create the entry basic block and finish
// the body at the end. To parse the expressions we will call the build.expr function for each expression in the file,
// as it is a function pointer it is not necessary to use the return value of builder.expr. Finally, we just added the
// function containing all the parsed expressions in the members of the file. This function should only be executed if
// the current file contains expressions, or an unnecessary temporary func will be created.
//
// Following is an example of code and the IR generated
//
// const express = require('express')
//
// Source Code:
//
// const app = express()
//
// app.get('/', (req, res) => {
//     console.log(req, res)
// });
//
// IR:
//
// func %fn2():
// 0:                                        entry
//       %t0 = make closure %fn2$1
//       %t1 = app.get("/", %t0)
//       %t2 = console.log("test")
func (f *File) buildExpressions() {
	var b builder

	fn := f.NewFunction(fmt.Sprintf("%%fn%d", len(f.Members)), f.syntax)
	fn.currentBlock = fn.newBasicBlock("entry")

	for _, expr := range f.expressions {
		b.expr(fn, expr, true)
	}

	fn.finishBody()
	f.Members[fn.name] = fn
}

// Build builds all function members of a struct s.
//
// For every method of a struct a receiver parameter is added. This parameter
// is named self and added as the first index of method.Signature.Params slice
func (s *Struct) Build() {
	for _, method := range s.Methods {
		method.Build()

		// Since this parameter don't exist in source code, the syntax for this parameter is nil.
		p := &Parameter{
			node:     node{nil},
			name:     "self", // TODO: this name should be normalized
			Value:    s,
			Receiver: true,
			parent:   method,
		}

		method.Signature.Params = append([]*Parameter{p}, method.Signature.Params...)
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
		locals:  make(map[string]*Var),
		Preds:   make([]*BasicBlock, 0),
		Succs:   make([]*BasicBlock, 0),
	}

	fn.Blocks = append(fn.Blocks, b)

	return b
}

// emit add a new instruction on current basic block of function.
func (fn *Function) emit(instr Instruction) {
	fn.currentBlock.emit(instr)
}

// addNamedLocal creates a local variable, adds it to function fn and return it.
//
// Subsequent calls to fn.lookup(name) will return the same variable.
func (fn *Function) addNamedLocal(name string, value Value, syntax ast.Node) *Var {
	v := fn.newVar(name, value, syntax)
	fn.addVariable(v)
	fn.emit(v)
	return v
}

// addLocal create a new temporary variable, adds it to function fn and return it.
//
// The temporary name used is %tN where N is the current number of local variables
// on fn. The % prefix is added on variable name to avoid collisions.
func (fn *Function) addLocal(value Value, syntax ast.Node) *Var {
	return fn.addNamedLocal("", value, syntax)
}

// addVariable add variable v to the current basic block of fn.
func (fn *Function) addVariable(v *Var) {
	fn.currentBlock.locals[v.Label] = v
	fn.nLocals++

	// Just append if the variable is declared on source code.
	if v.Label != "" {
		fn.Locals = append(fn.Locals, v)
	}
}

// newVar create a new variable at the current basic block of fn with the given label and value.
func (fn *Function) newVar(label string, value Value, syntax ast.Node) *Var {
	return &Var{
		node:  node{syntax},
		name:  fmt.Sprintf("%%t%d", fn.nLocals),
		Label: label,
		Value: value,
		block: fn.currentBlock,
	}
}

// finishBody finalizes the function after IR code generation of its body.
func (fn *Function) finishBody() {
	fn.currentBlock = nil
}

// emit appends an instruction to the current basic block.
// If the instruction defines a Value, it is returned.
func (b *BasicBlock) emit(i Instruction) {
	b.Instrs = append(b.Instrs, i)
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

// stmt convert a statement s to a IR form.
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
	case *ast.IfStmt:
		// Create if them and done blocks.
		then := fn.newBasicBlock("if.then")
		done := fn.newBasicBlock("if.done")

		// If the if statement don't have an else branch, use the done block as else.
		// Otherwise create new if else block.
		els := done
		if stmt.Else != nil {
			els = fn.newBasicBlock("if.else")
		}

		// Emit if condition to function.
		b.cond(fn, stmt.Cond, then, els)

		// Set current block to if then block and them process the if body.
		fn.currentBlock = then
		b.stmt(fn, stmt.Body)

		// Emit a jump to done block.
		emitJump(fn, done)

		// If the if statement has a else branch, set the current block to else branch,
		// process the else branch statement and them emit a jump to done block.
		if stmt.Else != nil {
			fn.currentBlock = els
			b.stmt(fn, stmt.Else)
			emitJump(fn, done)
		}

		fn.currentBlock = done
	case *ast.ForStatement:
		b.forStmt(fn, stmt)
	case *ast.TryStmt:
		b.tryStatement(fn, stmt)
	case *ast.WhileStmt:
		b.whileStmt(fn, stmt)
	case *ast.SwitchStatement:
		b.switchStatement(fn, stmt)
	case *ast.BadNode:
		// Do nothing with bad nodes.
	default:
		unsupportedNode(stmt)
	}
}

// tryStatement parse the ast.TryStmt to it's IR representation.
//
// nolint:gocyclo // centralizes all the try statement parse, necessary complexity.
func (b *builder) tryStatement(fn *Function, stmt *ast.TryStmt) {
	// Create the 'try.then' and 'try.done' blocks.
	then := fn.newBasicBlock("try.then")
	done := fn.newBasicBlock("try.done")

	// set the 'try.then' block and process the try statement body.
	fn.currentBlock = then
	b.stmt(fn, stmt.Body)

	// In case the try statement don't have a finalizer, use done instead.
	finally := done

	// If the statement contains a finalizer, creates a new 'try.finally' basic block and parse the finalizer body.
	if stmt.Finalizer != nil {
		finally = fn.newBasicBlock("try.finally")
		fn.currentBlock = finally

		// parse the all the statements of the finalizer body.
		for _, s := range stmt.Finalizer.List {
			b.stmt(fn, s)
		}

		// since that after the finally statement the try statement is over, emits a jump to the done basic block.
		emitJump(fn, done)

		// if there's no catch clause but there's a finally statement, a jump is emitted to the 'try.finally' block
		// in the 'try.then' block.
		if len(stmt.CatchClause) == 0 {
			fn.currentBlock = then
			emitJump(fn, finally)
		}
	}

	// In case the try statement don't have any catch clause, use done instead.
	catch := done

	// If the try statement contains at least one catch clause a basic block named 'try.catch' will be created.
	// This block will contain all the conditions related to the catch clauses exceptions, and they possible jumps.
	// Also, a new jump is added to the 'try.then' block into to the new 'try.catch' block.
	// Ex of the 'try.then' block:
	//
	//  1:						 try.then
	// 		console.log('try body')
	// 		jump 2
	//
	// Ex of the 'try.catch' block:
	//
	//  2:						 try.catch
	// 		if ex goto 'try.catch.N' else 'try.finally'
	// 		if ex goto 'try.catch.N' else 'try.finally'
	//
	if len(stmt.CatchClause) > 0 {
		catch = fn.newBasicBlock("try.catch")
		fn.currentBlock = then
		emitJump(fn, catch)
	}

	// parse all the catch clauses in the try statement, for each different clause will create a new basic block in
	// the following pattern: 'try.catch.N'.
	// Ex:
	//
	// 3:						 try.catch.0
	// 	console.log(ex)
	// 	jump N
	//
	// 3:						 try.catch.1
	// 	console.log(ex)
	// 	jump N
	for i, c := range stmt.CatchClause {
		// creates a new catch basic block and parse the catch body
		catchBlock := fn.newBasicBlock(fmt.Sprintf("try.catch.%d", i))
		fn.currentBlock = catchBlock
		b.stmt(fn, c.Body)

		// checks if there's a finalizer in the try statement, if so, it's added a jump to the 'try.finally', in case
		// there's no finally statement, a jump to the 'try.done' block is added. After the jump, in both scenarios a
		// goto is added to the 'try.catch' block informing a new 'try.catch.N' possible flow.
		// Ex:
		//
		//  "if ex goto 'try.catch.N' else 'try.finally'" added when there's a finally statement
		//  "if ex goto 'try.catch.N' else 'try.done'" added when there's no finally statement
		if stmt.Finalizer != nil {
			emitJump(fn, finally)
			fn.currentBlock = catch
			b.cond(fn, c.Parameter, catchBlock, finally)
		} else {
			emitJump(fn, done)
			fn.currentBlock = catch
			b.cond(fn, c.Parameter, catchBlock, done)
		}
	}

	fn.currentBlock = done
}

// switchStatement parse the ast.SwitchStatement to it's IR representation, the idea is to treat it as a normal if
// condition.
//
// 0:                                                                       entry
//        %t0 = console.log("switch entry")
//        %t1 = "2"
//        %t4 = %t1 == "1"
//        if %t4 goto 3 else 2
// 1:                                                                       if.done
//        %t5 = console.log("switch done")
// 2:                                                                       if.else
//        %t2 = console.log("switch case default")
//        jump 1
// 3:                                                                       if.then
//        %t3 = console.log("switch case 1")
//        jump 1
// 4:                                                                       if.done
//
// TODO: In the future, it would be interesting to review this code in search of improvements and reduce the complexity.
// nolint:gocyclo // despite the complexity, the idea is to centralize all switch case handling here
func (b *builder) switchStatement(fn *Function, stmt *ast.SwitchStatement) {
	// get the function current block.
	previouslyBlock := fn.currentBlock

	// creates a new done block.
	done := fn.newBasicBlock("if.done")

	// separate the 'ast.SwitchCase' statements from the 'ast.SwitchDefault', also remove the possible bad nodes that
	// can be in the 'stmt.Body.List' slice.
	cases, defaultCase := b.getSwitchCasesAndDefault(stmt)

	// create a new basic block to represents the 'ast.SwitchDefault' statement
	defaultBlock := done

	// check if the switch statement contains a default case.
	if defaultCase != nil {
		defaultBlock = fn.newBasicBlock("if.else")
		fn.currentBlock = defaultBlock

		// parse the switch default case body.
		for _, v := range defaultCase.Body {
			b.stmt(fn, v)
		}

		// since that after the default case statement the switch has ended, a jump is emitted to the done block.
		emitJump(fn, done)

		// since there's a possibility of a switch statement contains only a default case, this condition is necessary
		// to treat this scenario.
		if cases == nil {
			fn.currentBlock = previouslyBlock
			emitJump(fn, defaultBlock)
		}
	}

	// creates a map to store the switch 'then' and 'done' blocks
	thenBlocks := make(map[int]*BasicBlock, len(cases))
	doneBlocks := make(map[int]*BasicBlock, len(cases))

	for i, c := range cases {
		// if it's the first iteration of the for, a 'then' and a 'done' block it's going to be created, for the next
		// iterations the blocks are already be created, so instead of creating we are going to use the existing ones.
		if i == 0 {
			thenBlocks[i] = fn.newBasicBlock("if.then")
			doneBlocks[i] = fn.newBasicBlock("if.done")
		}

		// set the  actual iteration 'then' block
		fn.currentBlock = thenBlocks[i]

		// parse the 'ast.SwitchCase' body
		for _, v := range c.Body {
			b.stmt(fn, v)
		}

		// set the current block as the actual iteration 'then' block and emmit a jump to the 'done' block.
		// This happens cause after matching the condition of a case, the switch statement is over, and we can jump
		// to the done block.
		fn.currentBlock = thenBlocks[i]
		emitJump(fn, done)

		// checks if it's the last case, is so, the case it's already has been processed and there's no more flows to
		// process. There is just one exception to this, that is when the switch statement contains just one case and
		// a new condition is created to validate this. After these validations, the for is ended.
		if len(cases) == i+1 {
			// validate if it's the only case in the statement, is so, creates a new CFG condition.
			if i == 0 {
				// set the current block as the entry block.
				fn.currentBlock = previouslyBlock

				// a new condition is created with the 'then' block of this iteration and the 'defaultBlock' as else.
				// If there's no default case, the 'else' of the condition will be the 'done' block.
				b.cond(fn, b.switchCondExpr(c.Position, stmt.Value, c.Cond), thenBlocks[i], defaultBlock)
			}

			break
		}

		// if it's the first case, the condition of the CFG needs to be created in the entry block.
		if i == 0 {
			fn.currentBlock = previouslyBlock
			b.cond(fn, b.switchCondExpr(c.Position, stmt.Value, c.Cond), thenBlocks[i], doneBlocks[i])
		}

		// the following steps creates the next iteration blocks, they are going to be stored in the 'thenBlocks'
		// and 'doneBlocks', since we need to write they condition of the CFG in this iteration and use them in the
		// next iteration to parse the body.
		expr := b.switchCondExpr(c.Position, stmt.Value, cases[i+1].Cond)

		// check if this iteration it's the next to last, if so creates just 'then' block of the next iteration.
		if len(cases)-1 == i+1 {
			thenBlocks[i+1] = fn.newBasicBlock("if.then")

			// set the current block as the actual iteration done block
			fn.currentBlock = doneBlocks[i]

			// a new condition is created with the next iteration 'then' block of this iteration and the 'defaultBlock'
			// as else. If there's no default case, the 'else' of the condition will be the 'done' block.
			b.cond(fn, expr, thenBlocks[i+1], defaultBlock)
		} else {
			// if it's not the next to last case from the switch, a new 'then' and 'done' block are created.
			// These blocks represent the next case of the switch statement, and the CFG condition is going to be
			// written the into the actual case 'done' block. Since these blocks are going to be used in the next
			// iteration, they are stored in the 'thenBlocks' and 'doneBlocks' maps.
			thenBlocks[i+1] = fn.newBasicBlock("if.then")
			doneBlocks[i+1] = fn.newBasicBlock("if.done")

			fn.currentBlock = doneBlocks[i]
			b.cond(fn, expr, thenBlocks[i+1], doneBlocks[i+1])
		}
	}

	fn.currentBlock = done
}

// switchCondExpr creates a new binary expression that checks if the left value is equal the right value.
// Since we represent the switch cases as normals ifs, we need to create the binary expression that represents
// the condition for each case.
func (b *builder) switchCondExpr(position ast.Position, left, right ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{
		Position: position,
		Left:     left,
		Op:       "==",
		Right:    right,
	}
}

// getSwitchCasesAndDefault separate the 'ast.SwitchCase' statements from the 'ast.SwitchDefault', also remove the
// possible bad nodes that can be in the 'stmt.Body.List' slice, for example commented cases from the switch.
// EX:
//    console.log("switch entry")
//    let foo = 'bar'
//
//    switch (foo) {
//        // case 'a':
//        //     console.log('bad node, should be ignored')
//        case 'b':
//            console.log('switch case 1, should be appended')
//        default:
//            console.log("switch case default, should be returned")
//    }
//
//    console.log("switch done")
//
func (b *builder) getSwitchCasesAndDefault(stmt *ast.SwitchStatement) (c []*ast.SwitchCase, d *ast.SwitchDefault) {
	for _, s := range stmt.Body.List {
		switch s := s.(type) {
		case *ast.SwitchDefault:
			d = s
		case *ast.SwitchCase:
			c = append(c, s)
		}
	}

	return
}

// expr lowers a single-result expression e to IR form and return the Value defined by the expression.
//
// The expand parameter informs if expr should expand the expression e to a temporary variable. If
// expand is true and the expression e lowers to an IR Value that is also an Instruction, a temporary
// variable will be created using the Value created from e as the variable value, and the returned
// Value will be the created temporary variable. Otherwise the IR Value created from e is returned.
//
// Note that Var node is an exception to the rule informed above, because it is already a variable.
//
// nolint: gocyclo // cyclomatic complexity is necessary for now.
func (b *builder) expr(fn *Function, e ast.Expr, expand bool) Value {
	var value Value

	switch expr := e.(type) {
	// Value's that are *not* Instruction's (Var is an exception, see the doc above)
	case *ast.BasicLit:
		return &Const{
			node:  node{e},
			Value: expr.Value,
		}
	case *ast.Ident:
		if v := b.lookup(fn, expr.Name); v != nil {
			return v
		}
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
	case *ast.SelectorExpr:
		return b.selectorExpr(fn, expr)
	case *ast.KeyValueExpr:
		return &HashMap{
			node:  node{expr},
			Key:   b.expr(fn, expr.Key, false /* expand */),
			Value: b.expr(fn, expr.Value, false /* expand */),
		}
	case *ast.IncExpr:
		// Convert a++ to a = a + 1
		value = &BinOp{
			node: node{expr},
			Op:   expr.Op[:1], // Convert ++/-- to +/-
			Left: b.lookup(fn, expr.Arg.Name),
			Right: &Const{
				node:  node{nil},
				Value: "1",
			},
		}
		return fn.addNamedLocal(expr.Arg.Name, value, expr)

	// Value's that are also Instruction's.
	case *ast.FuncLit:
		// Create an anonymous function using the parent function name
		// and the current the total of anonymouns functions as a name.
		value = b.funcLit(fn, fmt.Sprintf("%s$%d", fn.Name(), len(fn.AnonFuncs)+1), expr)
	case *ast.CallExpr:
		value = b.callExpr(fn, expr)
	case *ast.BinaryExpr:
		value = b.binaryExpr(fn, expr)
	case *ast.ObjectExpr:
		value = b.objectExpr(fn, expr)
	case *ast.SubscriptExpr:
		value = &Lookup{
			node:   node{expr},
			Object: b.expr(fn, expr.Object, true /* expand */),
			Index:  b.expr(fn, expr.Index, true /* expand */),
		}
	default:
		unsupportedNode(expr)
		return nil
	}

	if expand {
		return fn.addLocal(value, e)
	}

	return value
}

// whileStmt emits code to fn for a while statement block.
// 0:                                                                         entry
// 		...previous code before loop...
// 		jump 2
// 1:                                                                    while.body
// 		...body of while loop...
// 2:                                                                    while.cond
// 		if cond goto 1 else 3
// 3:                                                                    while.done
// 		...code after while loop...
//
// TODO(matheus): Improve the IR generation for incomplete blocks.
//
// Incomplete blocks are blocks that further predecessors will be added after processing
// the code inside the block. Since the code can use variable defined in predecessors
// blocks (that was not added yet) we can't create correctly phi values to these variables
// so in this case we consider the variable value for the predecessor block that was already
// processed at the this point. In this case, the while.cond block is an incomplete block because
// the while.body block is added as predecessor after issuing the code of while.cond block.
func (b *builder) whileStmt(fn *Function, stmt *ast.WhileStmt) {
	// Create the while body.
	body := fn.newBasicBlock("while.body")

	// Create the while condition, if the while statement don't have
	// a condition, use the while.body to recursively jump.
	cond := body

	if stmt.Cond != nil {
		// If while has a condition, create a while.cond block to jump
		// otherwise jump to while.body.
		cond = fn.newBasicBlock("while.cond")
	}

	// Jump and set current block to cond, could be while.body or while.cond blocks.
	emitJump(fn, cond)
	fn.currentBlock = cond

	// Create the while done block that holds code after the while statement.
	done := fn.newBasicBlock("while.done")

	if stmt.Cond != nil {
		// Emit the while condition if exists and set the current block
		// from while.cond to while.body.
		b.cond(fn, stmt.Cond, body, done)
		fn.currentBlock = body
	}

	// Emit the while body and emit a jump from while.body to while.cond to represent
	// the loop. Note that if while statement don't have a condition this emission will
	// be for the while.body again to represent the endless recursion.
	b.stmt(fn, stmt.Body)
	emitJump(fn, cond)

	// Set current block to while.done to further processing code after while statement.
	fn.currentBlock = done
}

// forStmt emits code to fn for a for statement block.
//
// 0:                                                                         entry
//         ...previous code before loop...
//         jump 2
// 1:                                                                      for.body
//         ...body of loop...
//         jump 2
// 2:                                                                      for.loop
//         if cond goto 1 else 3
// 3:                                                                      for.done
//         ...code after loop...
//
// nolint: funlen,gocyclo // For loops are complicated, we can improve this in the future.
func (b *builder) forStmt(fn *Function, stmt *ast.ForStatement) {
	// Create the body of for loop.
	body := fn.newBasicBlock("for.body")

	// If for statement has a condition, create a for loop block
	// to emit loop condition. Otherwise use the body of loop
	// to emit a jump.
	loop := body
	if stmt.Cond != nil {
		loop = fn.newBasicBlock("for.loop")
	}

	// Emit a jump from previous basic block to for.loop condition.
	emitJump(fn, loop)
	fn.currentBlock = loop

	phis := make([]*Phi, 0)

	// Emit variable declaration on for statement on for.loop (if stmt.Cond is not nil, otherwise
	// will emit on for.body) block.
	if stmt.VarDecl != nil {
		b.stmt(fn, stmt.VarDecl)

		// Since variables created at stmt.VarDecl could be changed at stmt.Increment
		// we should create phi-nodes to these variables created. The phi-node created
		// here will only contain a single edge that is the variable created at stmt.VarDecl
		// if this variable is also changed in stmt.Increment we should append this change
		// on edges of the phi-node created here. This is necessary to represent a multiple
		// possible values of a increment variable inside a for loop, for example:
		// for (i = 0; i < len(data); i++) {}
		// The variable i above can have two possible values; 0 if len(data) == 0 or
		// i = i + 1, since this variable is incremented at every iteration.
		for name, v := range loop.locals {
			// Only check for variable declared on source code.
			if v.Label != "" {
				phi := &Phi{
					Comment: v.Label,
					Edges:   []*Var{v},
				}
				phis = append(phis, phi)

				fn.addNamedLocal(name, phi, nil)
			}
		}
	}

	// Create for.done block to jump if for statement has a condition.
	done := fn.newBasicBlock("for.done")

	// Emit for loop condition to fn if exists.
	if stmt.Cond != nil {
		b.cond(fn, stmt.Cond, body, done)
		fn.currentBlock = body
	}

	// Emit increment on for.body condition.
	b.expr(fn, stmt.Increment, true /*expand*/)

	// Here we check if the edges of phi-node created above has a change on
	// stmt.Increment block, if has, we append this change as a new edge on
	// phi-node.
	for _, phi := range phis {
		for _, edge := range phi.Edges {
			if v, exists := body.locals[edge.Label]; exists {
				phi.Edges = append(phi.Edges, v)
			}
		}
	}

	// Emit the for body on loop.body block.
	b.stmt(fn, stmt.Body)

	// Emit a jump from for.body to to for.loop again.
	emitJump(fn, loop)

	// Set the current block to for.done to finish the for loop
	fn.currentBlock = done
}

// selectorExpr return the Value of a selector expression.
//
// The value returned will be a selector expression entire resolve. If the selector
// expression contains call expressions (or any expression that can be expanded), these
// expressions will be separated and temp variables will be created to store these call
// expressions and these local variables will be used on the resolve selector expression
// string.
//
// Example:
// Source:
//	a.b.c()
// IR:
//	%t0 = a.b.c()
//
// Source:
//	a.b().c.d()
// IR:
//	%t0 = a.b()
//	%t1 = %t0.c.d()
//
// Selector expressions that represents methods calls are printed using the object type
// instead the variable name:
//
// Source:
//	let foo = new Foo()
//	foo.something()
// IR:
//	%t0 = constructor(Foo)
//	%t1 = Foo.something()
func (b *builder) selectorExpr(fn *Function, expr *ast.SelectorExpr) Value {
	var value Value

	switch e := expr.Expr.(type) {
	case *ast.Ident:
		if v := b.lookup(fn, e.Name); v != nil {
			// e.Name could be an identifier to a declared Value, that could be a
			// a variable or a imported member so we lookup to get the correctly name.
			value = v
		} else {
			// Otherwise we just use the identifier name, since we can't find any
			// reference to this identifier.
			value = &Var{
				node:  node{e},
				name:  e.Name,
				Label: e.Name,
				Value: nil,
			}
		}
	case *ast.SelectorExpr:
		value = b.selectorExpr(fn, e)
	default:
		value = b.expr(fn, e, true /*expand */)
	}

	return &Selector{
		node:  node{expr},
		Value: value,
		// Since we can lookup the variable from value context we just create
		// a variable with a nil value.
		Field: &Var{
			node:  node{expr.Sel},
			name:  expr.Sel.Name,
			Value: nil,
		},
	}
}

// cond emits to fn code to evaluate boolean condition e and jump
// to t(true) or f(false) depending on its value.
func (b *builder) cond(fn *Function, e ast.Expr, t, f *BasicBlock) {
	emitIf(fn, b.expr(fn, e, true /*expand */), t, f)
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
		b.assignValue(fn, lhs, rhs, syntax)
	case *ast.SelectorExpr:
		fn.addNamedLocal(b.selectorExpr(fn, lhs).Name(), b.expr(fn, rhs, false /*expand*/), syntax)
	default:
		unsupportedNode(lhs)
	}
}

// assignValue create a new Value according to the Value yielded from rhs expression.
//
// If the rhs expression is a *ast.FuncLit a new Closure Instruction will be created and emitted
// to fn. If the rhs expression is an identifier, the identifier value will be resolved to set
// the identifier value as the value of lhs identifier. Otherwise, a new named local variable
// will be created using the Value created from rhs expression.
func (b *builder) assignValue(fn *Function, lhs *ast.Ident, rhs ast.Expr, syntax *ast.AssignStmt) {
	switch rhs := rhs.(type) {
	case *ast.FuncLit:
		fn.emit(b.funcLit(fn, lhs.Name, rhs))
	case *ast.Ident:
		fn.addNamedLocal(lhs.Name, b.lookup(fn, rhs.Name), syntax)
	case *ast.ObjectExpr:
		b.objectExprFromAssign(fn, lhs, rhs, syntax)
	default:
		fn.addNamedLocal(lhs.Name, b.expr(fn, rhs, false /*expand*/), syntax)
	}
}

// objectExprFromAssign parse rhs assign when it's an ast.ObjectExpr, creating a new ir.Object that can represent an
// array, constructor or hashmap. After the parse the value is added to the function locals.
func (b *builder) objectExprFromAssign(fn *Function, lhs *ast.Ident, rhs *ast.ObjectExpr, syntax *ast.AssignStmt) {
	obj := b.objectExpr(fn, rhs)

	fn.addNamedLocal(lhs.Name, obj, syntax)
}

// objectExpr create a new ir.Object from ast.ObjectExpr expr. The ir.Object can represent an array, constructor
// or hashmap.
func (b *builder) objectExpr(fn *Function, expr *ast.ObjectExpr) *Object {
	obj := &Object{
		node:    node{expr},
		Comment: expr.Comment,
	}

	if expr.Type != nil {
		obj.Type = b.expr(fn, expr.Type, false /*expand*/)
	}

	for _, expr := range expr.Elts {
		obj.Values = append(obj.Values, b.expr(fn, expr, true /*expand*/))
	}

	return obj
}

// lookup return the Value declared on source file with the given name. The search
// order is; first check at function level, them function signature, them global values
// on file and finally for a imported value.
//
// nolint: gocyclo // We need to lookup the value declaration in to many places.
func (b *builder) lookup(fn *Function, name string) Value {
	if v := fn.lookup(name); v != nil {
		return v
	}

	for _, param := range fn.Signature.Params {
		if param.name == name {
			return param
		}
	}

	if global, ok := fn.File.Members[name].(*Global); ok {
		return fn.addLocal(b.expr(fn, global.Value, false /* expand */), global.Value)
	}

	if importt := fn.File.ImportedPackage(name); importt != nil {
		return importt
	}

	return nil
}

// stmtList emits to fn code for all statements in list.
func (b *builder) stmtList(fn *Function, list []ast.Stmt) {
	for _, s := range list {
		b.stmt(fn, s)
	}
}

// funcLit crate a new Closure to a given AST based function literal.
func (b *builder) funcLit(parent *Function, name string, syntax *ast.FuncLit) *Closure {
	fn := parent.File.NewFunction(name, syntax)
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
// call arguments will point to to this declared variable.
//
// nolint:gocyclo // Some checks is needed here.
func (b *builder) callExpr(parent *Function, call *ast.CallExpr) *Call {
	args := make([]Value, 0, len(call.Args))

	for _, arg := range call.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			// If identifier used on function call is declared inside the parent function
			// we use this declared variable as argument to function call, otherwise if
			// just parse the expression value.
			if local := b.lookup(parent, ident.Name); local != nil {
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
		fn.name = b.selectorExpr(parent, call).Name()
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
			// Since default paramenter values can not have more than
			// one value, we check if the value really exists and use
			// to create the parameter value.
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

// addEdge adds a control-flow graph edge from f to t.
func addEdge(f, t *BasicBlock) {
	f.Succs = append(f.Succs, t)
	t.Preds = append(t.Preds, f)
}

// emitJump emits to fn a jump to target, and updates the control-flow graph.
func emitJump(fn *Function, target *BasicBlock) {
	b := fn.currentBlock
	b.emit(&Jump{b})
	addEdge(b, target)
	fn.currentBlock = nil
}

// emitIf emits to fn a conditional jump to tblock or fblock based on
// cond, and updates the control-flow graph.
func emitIf(fn *Function, cond Value, tblock, fblock *BasicBlock) {
	b := fn.currentBlock
	b.emit(&If{
		Cond:  cond,
		block: b,
	})
	addEdge(b, tblock)
	addEdge(b, fblock)
	fn.currentBlock = nil
}
