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
	"bytes"
	"fmt"

	"github.com/ZupIT/horusec-engine/internal/cst"
)

// nosec is the directive to ignore a block of code.
var nosec = []byte("#nosec")

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
// Position implements Node.Pos method to easily embedded on Node implementations,
// so Node implementations can just inherit Position and these Nodes will fully
// implement the Node interface.
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
			Row:    start.Row + 1, // tree-sitter row start at 0.
			Column: start.Column,
		},
		end: Pos{
			Byte:   n.EndByte(),
			Row:    end.Row + 1, // tree-sitter row start at 0.
			Column: end.Column,
		},
	}
}

// Start return the start position.
func (p Position) Start() Pos { return p.start }

// End return the end position.
func (p Position) End() Pos { return p.end }

// Start implements Node.Pos.
func (p Position) Pos() Position { return p }

// Node represents a generic node on AST.
//
// There are 3 main classes of nodes: Expressions, Statements, and
// Declaration nodes. The node fields correspond to the individual
// parts of the respective productions.
//
// All node types implement the Node interface.
type Node interface {
	Pos() Position
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

// BadNode node is a placeholder for a syntax errors or syntaxes that
// is no supported yet.
//
// BadNode implements Decl, Expr and Stmt so can be used in any place.
type BadNode struct {
	Position
	Comment string // Optional comment for debugging.
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
		Name    *Ident // Name of identifier or nil.
		Type    Expr   // Type of object or nil.
		Elts    []Expr // Elements used to created object or nil.
		Comment string // Optional type of the object expression; no semantic significance
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

	// IncExpr node represents a variable increment expression
	IncExpr struct {
		Position
		Op  string // Operator. // TODO: This should be a concrete type.
		Arg *Ident // identifier of the argument being incremented
	}

	// SubscriptExpr node represents a subscript expression
	SubscriptExpr struct {
		Position
		Object Expr // array object
		Index  Expr // index of the subscript
	}
)

func (*Ident) expr()         {}
func (*Field) expr()         {}
func (*FieldList) expr()     {}
func (*BasicLit) expr()      {}
func (*BinaryExpr) expr()    {}
func (*CallExpr) expr()      {}
func (*SelectorExpr) expr()  {}
func (*ObjectExpr) expr()    {}
func (*KeyValueExpr) expr()  {}
func (*FuncType) expr()      {}
func (*FuncLit) expr()       {}
func (*TemplateExpr) expr()  {}
func (*IncExpr) expr()       {}
func (*BadNode) expr()       {}
func (*SubscriptExpr) expr() {}

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

	// SwitchStatement node represents a switch statement
	SwitchStatement struct {
		Position
		Value Expr       // Value to switch case
		Body  *BlockStmt // Switch body
	}

	// SwitchCase node represents a single switch case
	SwitchCase struct {
		Position
		Cond Expr   // Condition that needs to match to enter this case
		Body []Stmt // Switch case body or nil
	}

	// SwitchDefault node represents a switch default statement
	SwitchDefault struct {
		Position
		Body []Stmt // Switch default case body
	}

	// BreakStatement node represents a break statement
	BreakStatement struct {
		Position
		Label *Ident // Break statement label
	}

	// ForStatement node represents a for statement
	ForStatement struct {
		Position
		VarDecl   Stmt       // For initializer
		Cond      Expr       // Condition needed to match
		Increment Expr       // For Increment expression
		Body      *BlockStmt // For body
	}

	// ForInStatement node represents a for in statement
	ForInStatement struct {
		Position
		Left  Expr       // Value on the left side of for in containing the value of the current iteration
		Right Expr       // Value on the right side of for in containing an array to iterate over
		Body  *BlockStmt // For in body
	}
	// ContinueStatement node represents a continue statement
	ContinueStatement struct {
		Position
		Label *Ident // Continue statement label
	}

	// LabeledStatement node represents a labeled statement
	LabeledStatement struct {
		Position
		Label *Ident // LabeledStatement label
		Body  Stmt
	}
)

func (*BlockStmt) stmt()         {}
func (*AssignStmt) stmt()        {}
func (*ExprStmt) stmt()          {}
func (*ReturnStmt) stmt()        {}
func (*IfStmt) stmt()            {}
func (*TryStmt) stmt()           {}
func (*CatchClause) stmt()       {}
func (*WhileStmt) stmt()         {}
func (*SwitchStatement) stmt()   {}
func (*SwitchCase) stmt()        {}
func (*SwitchDefault) stmt()     {}
func (*BreakStatement) stmt()    {}
func (*ForStatement) stmt()      {}
func (*ContinueStatement) stmt() {}
func (*LabeledStatement) stmt()  {}
func (*ForInStatement) stmt()    {}
func (*BadNode) stmt()           {}

// ----------------------------------------------------------------------------
// Declarations
//
// A declaration is represented by one of the following declarations nodes.
//
type (
	// ImportDecl node represents a single package/module import.
	ImportDecl struct {
		Position
		Name  *Ident // Import name.
		Alias *Ident // Alias name or nil.
		Path  *Ident // Import path.
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
func (*BadNode) decl()    {}

// NewUnsupportedNode create a new BadNode for unsupported cst nodes.
func NewUnsupportedNode(n *cst.Node) *BadNode {
	return &BadNode{
		Position: NewPosition(n),
		Comment:  fmt.Sprintf("unsupported node type <%s>", n.Type()),
	}
}

// File node represents a program source file.
type File struct {
	Position
	Name     *Ident // file name.
	Decls    []Decl // top-level declarations or nil.
	Exprs    []Expr // top-level expressions or nil.
	BadNodes []Node // top-level unsupported nodes.
}

func NewIdent(node *cst.Node) *Ident {
	return &Ident{
		Name:     string(cst.SanitizeNodeValue(node.Value())),
		Position: NewPosition(node),
	}
}

// IsNosec return true if the given slice of bytes contains nosec directive.
func IsNosec(s []byte) bool {
	return bytes.Contains(s, nosec)
}

func (n *BadNode) String() string {
	pos := n.Pos().start
	return fmt.Sprintf("%s %d:%d", n.Comment, pos.Row, pos.Column)
}
