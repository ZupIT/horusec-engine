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

import "fmt"

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(node Node) (w Visitor)
}

// Inspect traverses an AST in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a
// call of f(nil).
func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}

// Walk traverses an AST in depth-first order: It starts by calling
// v.Visit(node); node must not be nil. If the visitor w returned by
// v.Visit(node) is not nil, Walk is invoked recursively with visitor
// w for each of the non-nil children of node, followed by a call of
// w.Visit(nil).
//
// nolint:funlen,gocognit,gocyclo // To many type checks to do.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	// walk children
	// (the order of the cases matches the order
	// of the corresponding node types in ast.go)
	switch n := node.(type) {
	// Expressions
	case *Ident, *BasicLit:
		// Nothing to do.
	case *Field:
		if n.Name != nil {
			Walk(v, n.Name)
		}
	case *FieldList:
		for _, f := range n.List {
			Walk(v, f)
		}
	case *FuncType:
		if n.Params != nil {
			Walk(v, n.Params)
		}
		if n.Results != nil {
			Walk(v, n.Results)
		}
	case *FuncLit:
		Walk(v, n.Type)
		Walk(v, n.Body)
	case *TemplateExpr:
		for _, expr := range n.Subs {
			Walk(v, expr)
		}
	case *ObjectExpr:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.Type != nil {
			Walk(v, n.Type)
		}
		walkExprList(v, n.Elts)
	case *BinaryExpr:
		Walk(v, n.Left)
		Walk(v, n.Right)
	case *SelectorExpr:
		Walk(v, n.Expr)
		if n.Sel != nil {
			Walk(v, n.Sel)
		}
	case *CallExpr:
		Walk(v, n.Fun)
		walkExprList(v, n.Args)
	case *KeyValueExpr:
		Walk(v, n.Key)
		Walk(v, n.Value)
	case *IncExpr:
		if n.Arg != nil {
			Walk(v, n.Arg)
		}

	// Statements
	case *AssignStmt:
		walkExprList(v, n.LHS)
		walkExprList(v, n.RHS)
	case *BlockStmt:
		walkStmtList(v, n.List)
	case *ExprStmt:
		Walk(v, n.Expr)
	case *ReturnStmt:
		walkExprList(v, n.Results)
	case *IfStmt:
		Walk(v, n.Cond)
		Walk(v, n.Body)
		if n.Else != nil {
			Walk(v, n.Else)
		}
	case *WhileStmt:
		Walk(v, n.Cond)
		if n.Body != nil {
			Walk(v, n.Body)
		}
	case *TryStmt:
		if n.Body != nil {
			Walk(v, n.Body)
		}

		for _, value := range n.CatchClause {
			Walk(v, value)
		}

		if n.Finalizer != nil {
			Walk(v, n.Finalizer)
		}
	case *SwitchStatement:
		if n.Value != nil {
			Walk(v, n.Value)
		}

		if n.Body != nil {
			Walk(v, n.Body)
		}
	case *SwitchCase:
		if n.Cond != nil {
			Walk(v, n.Cond)
		}

		walkStmtList(v, n.Body)
	case *SwitchDefault:
		walkStmtList(v, n.Body)
	case *ForStatement:
		if n.VarDecl != nil {
			Walk(v, n.VarDecl)
		}

		if n.Cond != nil {
			Walk(v, n.Cond)
		}

		if n.Increment != nil {
			Walk(v, n.Increment)
		}

		if n.Body != nil {
			Walk(v, n.Body)
		}
	case *ForInStatement:
		if n.Left != nil {
			Walk(v, n.Left)
		}

		if n.Right != nil {
			Walk(v, n.Right)
		}

		if n.Body != nil {
			Walk(v, n.Body)
		}
	// Declarations
	case *ImportDecl:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		if n.Path != nil {
			Walk(v, n.Path)
		}
	case *ValueDecl:
		walkIdentList(v, n.Names)
		walkExprList(v, n.Values)
	case *FuncDecl:
		Walk(v, n.Name)
		Walk(v, n.Type)
		if n.Body != nil {
			Walk(v, n.Body)
		}
	case *ClassDecl:
		Walk(v, n.Name)
		walkDeclList(v, n.Body.List)

	// Files
	case *File:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		walkDeclList(v, n.Decls)
		walkExprList(v, n.Exprs)

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}
}

func walkDeclList(v Visitor, list []Decl) {
	for _, x := range list {
		Walk(v, x)
	}
}

func walkIdentList(v Visitor, list []*Ident) {
	for _, x := range list {
		Walk(v, x)
	}
}

func walkStmtList(v Visitor, list []Stmt) {
	for _, x := range list {
		Walk(v, x)
	}
}

func walkExprList(v Visitor, list []Expr) {
	for _, x := range list {
		Walk(v, x)
	}
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}

	return nil
}
