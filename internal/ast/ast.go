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
	"fmt"

	"github.com/ZupIT/horusec-engine/internal/cst"
)

// Pos represents a start or end position of a node on source code.
type Pos struct {
	Byte   uint32 // Start or end byte of a node on source code.
	Row    uint32 // Start or end row of a node on source file.
	Column uint32 // Start or end column of a node on source file.
}

func (p Pos) String() string {
	return fmt.Sprintf("byte=%d, row=%d, column=%d", p.Byte, p.Row, p.Column)
}

// Position wraps Pos with a start and end position of a node.
//
// Position implements Node.Start and Node.End methods to easily embedded
// on Node implementations, so Node implementations can just inherit Position
// and these Nodes will fully implement the Node interface.
type Position struct {
	start Pos
	end   Pos
}

// NewPosition create a new Position from cst.Node.
func NewPosition(n *cst.Node) Position {
	start := n.StartPoint()
	end := n.EndPoint()

	return Position{
		start: Pos{
			Byte:   n.StartByte(),
			Row:    start.Row,
			Column: start.Column,
		},
		end: Pos{
			Byte:   n.EndByte(),
			Row:    end.Row,
			Column: end.Column,
		},
	}
}

// Start implements Node.Start.
func (p Position) Start() Pos { return p.start }

// End implements Node.End.
func (p Position) End() Pos { return p.end }

// Node represents a generic node on AST.
//
// There are 3 main classes of nodes: Expressions, Statements, and
// Declaration nodes. The node fields correspond to the individual
// parts of the respective productions.
//
// All node types implement the Node interface.
type Node interface {
	Start() Pos
	End() Pos
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
		Position
	}

	// BasicLit node represents a literal of basic type.
	BasicLit struct {
		Position
		Kind  string // TODO: This should be a concrete type
		Value string // literal string; e.g. 42, 0x7f, 3.14, 1e-9, 2.4i, 'a', '\x7f', "foo" or `\m\n\o`
	}

	// Field node represents a Field declaration in a function parameters/result.
	Field struct {
		Position
		Name Expr // Expression of field.
	}

	// FieldList represents a list of Fields.
	FieldList struct {
		Position
		List []*Field // Field list or nil.
	}

	// FuncType node represents a function type.
	FuncType struct {
		Position
		Params  *FieldList // Parameters of function; non nil.
		Results *FieldList // Results of function or nil.
	}

	// FuncLit node represents a function literal.
	FuncLit struct {
		Position
		Type *FuncType  // Function signature.
		Body *BlockStmt // Function body.
	}

	// TemplateExpr node represents a template string.
	TemplateExpr struct {
		Position
		Value string // Template string.
		Subs  []Expr // Substitution expressions.
	}

	// ObjectExpr node represents a object creation expression.
	ObjectExpr struct {
		Position
		Name *Ident // Name of identifier or nil.
		Type Expr   // Type of object or nil.
		Elts []Expr // Elements used to created object or nil.
	}

	// BinaryExpr node represents a binary expression.
	BinaryExpr struct {
		Position
		Left  Expr   // Left operand.
		Op    string // Operator. // TODO: This should be a concrete type.
		Right Expr   // Right operand.
	}

	// SelectorExpr node represents an expression followed by a selector.
	SelectorExpr struct {
		Position
		Expr Expr   // Expression
		Sel  *Ident // Field selector
	}

	// CallExpr node represents an expression followed by an argument list.
	CallExpr struct {
		Position
		Fun  Expr   // Function expression.
		Args []Expr // Function arguments or nil.
	}

	// KeyValueExpr node represents (key: value) pairs.
	KeyValueExpr struct {
		Position
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
		Position
		LHS []Expr
		RHS []Expr
	}

	// BlockStmt node represents a statement list.
	BlockStmt struct {
		Position
		List []Stmt // List of statements withim a block.
	}

	// ExprStmt node represents a (stand-alone) expression in a statement list.
	ExprStmt struct {
		Position
		Expr Expr // Expression
	}

	// ReturnStmt node represents a return statement.
	ReturnStmt struct {
		Position
		Results []Expr // result expressions; or nil
	}

	// IfStmt node represents an if statement.
	IfStmt struct {
		Position
		Cond Expr       // condition
		Body *BlockStmt // body of if branch
		Else Stmt       // else branch; or nil
	}

	// TryStmt node represents a try catch finally block
	TryStmt struct {
		Position
		Body        *BlockStmt     // Body of try
		CatchClause []*CatchClause // List of catches or nil
		Finalizer   *BlockStmt     // Block of finally statement or nil
	}

	// CatchClause node represents a catch inside a try catch statement
	CatchClause struct {
		Position
		Parameter *Ident     // Parameter inside catch block
		Body      *BlockStmt // Body of catch clause
	}

	// WhileStmt node represents a while block.
	WhileStmt struct {
		Position
		Cond Expr       // While loop condition.
		Body *BlockStmt // Body of while.
	}
)

func (*BlockStmt) stmt()   {}
func (*AssignStmt) stmt()  {}
func (*ExprStmt) stmt()    {}
func (*ReturnStmt) stmt()  {}
func (*IfStmt) stmt()      {}
func (*TryStmt) stmt()     {}
func (*CatchClause) stmt() {}
func (*WhileStmt) stmt()   {}

func (*BlockStmt) node()   {}
func (*AssignStmt) node()  {}
func (*ExprStmt) node()    {}
func (*ReturnStmt) node()  {}
func (*IfStmt) node()      {}
func (*TryStmt) node()     {}
func (*CatchClause) node() {}
func (*WhileStmt) node()   {}

// ----------------------------------------------------------------------------
// Declarations
//
// A declaration is represented by one of the following declarations nodes.
//
type (
	// ImportDecl node represents a single package/module import.
	ImportDecl struct {
		Position
		Name *Ident // Import name.
		Path *Ident // Import path.
	}

	// ValueDecl node represents a constant or variable declaration.
	ValueDecl struct {
		Position
		Names  []*Ident // Value names.
		Values []Expr   // Initial values or nil.
	}

	// FuncDecl node represents a function declaration.
	FuncDecl struct {
		Position
		Name *Ident     // Function name.
		Type *FuncType  // Function signature
		Body *BlockStmt // Function body.
	}

	// BodyDecl node represents a body declaration.
	BodyDecl struct {
		Position
		List []Decl // List of declarations inside body.
	}

	// ClassDecl node represents a class declaration.
	ClassDecl struct {
		Position
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
	Position
	Name  *Ident // file name.
	Decls []Decl // top-level declarations or nil.
	Exprs []Expr // top-level expressions or nil.
}

func (f *File) node() {}

func NewIdent(node *cst.Node) *Ident {
	return &Ident{
		Name:     string(cst.SanitizeNodeValue(node.Value())),
		Position: NewPosition(node),
	}
}
