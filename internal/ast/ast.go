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
	"github.com/ZupIT/horusec-engine/internal/cst"
)

// Node represents a generic node on AST.
//
// There are 3 main classes of nodes: Expressions, Statements, and
// Declaration nodes. The node fields correspond to the individual
// parts of the respective productions.
//
// All node types implement the Node interface.
type Node interface {
	node()
}

// Decl represents declaration nodes.
type Decl interface {
	Node
	decl()
}

// Expr represents expression nodes.
type Expr interface {
	Node
	expr()
}

// Stmt represents statement nodes.
type Stmt interface {
	Node
	stmt()
}

// ----------------------------------------------------------------------------
// Expressions
//
// An expression is represented by one of the following declarations nodes.
//
type (
	// Ident node represents an identifier.
	Ident struct {
		Name string
	}

	// BasicLit node represents a literal of basic type.
	BasicLit struct {
		Kind  string // TODO: This should be a concrete type
		Value string // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// Field node represents a Field declaration in a function parameters/result.
	Field struct {
		Name Expr // Expression of field.
	}

	// FieldList represents a list of Fields.
	FieldList struct {
		List []*Field // Field list or nil.
	}

	// FuncType node represents a function type.
	FuncType struct {
		Params  *FieldList // Parameters of function; non nil.
		Results *FieldList // Results of function or nil.
	}

	// FuncLit node represents a function literal.
	FuncLit struct {
		Type *FuncType  // Function signature.
		Body *BlockStmt // Function body.
	}

	// TemplateExpr node represents a template string.
	TemplateExpr struct {
		Value string // Template string.
		Subs  []Expr // Substitution expressions.
	}

	// ObjectExpr node represents a object creation expression.
	ObjectExpr struct {
		Name *Ident // Name of identifier or nil.
		Type Expr   // Type of object or nil.
		Elts []Expr // Elements used to created object or nil.
	}

	// BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		Left  Expr   // Left operand.
		Op    string // Operator. // TODO: This should be a concrete type.
		Right Expr   // Right operand.
	}

	// SelectorExpr node represents an expression followed by a selector.
	SelectorExpr struct {
		Expr Expr   // Expression
		Sel  *Ident // Field selector
	}

	// CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Fun  Expr   // Function expression.
		Args []Expr // Function arguments or nil.
	}

	// KeyValueExpr node represents (key: value) pairs.
	KeyValueExpr struct {
		Key   Expr
		Value Expr
	}
)

func (*Ident) expr()        {}
func (*Field) expr()        {}
func (*FieldList) expr()    {}
func (*BasicLit) expr()     {}
func (*BinaryExpr) expr()   {}
func (*CallExpr) expr()     {}
func (*SelectorExpr) expr() {}
func (*ObjectExpr) expr()   {}
func (*KeyValueExpr) expr() {}
func (*FuncType) expr()     {}
func (*FuncLit) expr()      {}
func (*TemplateExpr) expr() {}

func (*Ident) node()        {}
func (*Field) node()        {}
func (*FieldList) node()    {}
func (*BasicLit) node()     {}
func (*BinaryExpr) node()   {}
func (*CallExpr) node()     {}
func (*SelectorExpr) node() {}
func (*ObjectExpr) node()   {}
func (*KeyValueExpr) node() {}
func (*FuncType) node()     {}
func (*FuncLit) node()      {}
func (*TemplateExpr) node() {}

// ----------------------------------------------------------------------------
// Statements
//
// A statement is represented by one of the following declarations nodes
//
type (
	// AssignStmt node represents an assignment or a variable declaration.
	AssignStmt struct {
		Lhs []Expr
		Rhs []Expr
	}

	// BlockStmt node represents a statement list.
	BlockStmt struct {
		List []Stmt // List of statements withim a block.
	}

	// ExprStmt node represents a (stand-alone) expression in a statement list.
	ExprStmt struct {
		Expr Expr // Expression
	}

	// ReturnStmt node represents a return statement.
	ReturnStmt struct {
		Results []Expr // result expressions; or nil
	}

	// IfStmt node represents an if statement.
	IfStmt struct {
		Cond Expr       // condition
		Body *BlockStmt // body of if branch
		Else Stmt       // else branch; or nil
	}
)

func (*BlockStmt) stmt()  {}
func (*AssignStmt) stmt() {}
func (*ExprStmt) stmt()   {}
func (*ReturnStmt) stmt() {}
func (*IfStmt) stmt()     {}

func (*BlockStmt) node()  {}
func (*AssignStmt) node() {}
func (*ExprStmt) node()   {}
func (*ReturnStmt) node() {}
func (*IfStmt) node()     {}

// ----------------------------------------------------------------------------
// Declarations
//
// A declaration is represented by one of the following declarations nodes.
//
type (
	// ImportDecl node represents a single package/module import.
	ImportDecl struct {
		Name *Ident // Import name.
		Path *Ident // Import path.
	}

	// ValueDecl node represents a constant or variable declaration.
	ValueDecl struct {
		Names  []*Ident // Value names.
		Values []Expr   // Initial values or nil.
	}

	// FuncDecl node represents a function declaration.
	FuncDecl struct {
		Name *Ident     // Function name.
		Type *FuncType  // Function signature
		Body *BlockStmt // Function body.
	}

	// BodyDecl node represents a body declaration.
	BodyDecl struct {
		List []Decl // List of declarations inside body.
	}

	// ClassDecl node represents a class declaration.
	ClassDecl struct {
		Name *Ident    // Class name.
		Body *BodyDecl // Class body.
	}
)

func (*ImportDecl) decl() {}
func (*FuncDecl) decl()   {}
func (*ValueDecl) decl()  {}
func (*ClassDecl) decl()  {}
func (*BodyDecl) decl()   {}

func (*ImportDecl) node() {}
func (*FuncDecl) node()   {}
func (*ValueDecl) node()  {}
func (*ClassDecl) node()  {}
func (*BodyDecl) node()   {}

// File node represents a program source file.
type File struct {
	Name  *Ident // file name.
	Decls []Decl // top-level declarations or nil.
	Exprs []Expr // top-level expressions or nil.
}

func (f *File) node() {}

func NewIdent(node *cst.Node) *Ident {
	return &Ident{
		Name: string(cst.SanitizeNodeValue(node.Value())),
	}
}
