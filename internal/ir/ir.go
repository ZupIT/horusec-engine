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

package ir

import (
	"fmt"

	"github.com/ZupIT/horusec-engine/internal/ast"
)

// Member is a member of a file, like functions, global variables and constants.
type Member interface {
	member()

	// Name returns the declared name of the package member
	Name() string
}

// Value is an IR value that can be referenced by an instruction.
type Value interface {
	value()

	// Name returns the name of this value, and determines how
	// this Value appears when used as an operand of an
	// Instruction.
	//
	// This is the same as the source name for Parameters,
	// Functions and Vars.
	// For constants, it is a representation of the constant's value
	Name() string

	// String returns its disassembled form if this value is an Instruction;
	// otherwise it returns the name of the Value.
	String() string
}

// Instruction is an IR instruction that computes a new Value or has some effect.
//
// An Instruction that defines a value also implements the Value interface, an
// Instruction that only has an effect does not. For example, Var and Call
// implement Instruction and Value interfaces because both can define a value;
// Var define a new variable value and the call can set a value with its return.
type Instruction interface {
	instr()

	// String returns the disassembled form of this value.
	String() string
}

// File represents a single file converted to IR representation.
type File struct {
	name    string            // Name of file.
	Members map[string]Member // All file members keyed by name.

	expresions []ast.Expr                 // Top level file expressions.
	imported   map[string]*ExternalMember // All importable packages, keyed by import path.
}

// ExternalMember represents a member that is declared outside the file that is being used.
//
// ExternalMember is created from imports of a single file.
type ExternalMember struct {
	name  string // Named import member.
	Path  string // Full import path of member.
	Alias string // Alias declared on import; or empty.
}

// BasicBlock represents an IR basic block.
type BasicBlock struct {
	Index   int           // Index of this block on Function that this block belongs.
	Comment string        // Optional label; no semantic significance
	Instrs  []Instruction // Instructions in order.
}

// Function represents a function or method with the parameters and signature.
type Function struct {
	name      string          // Function name.
	File      *File           // File that this function belongs.
	Signature *Signature      // Function signature.
	Locals    map[string]*Var // Local variables of this function.
	Blocks    []*BasicBlock   // Basic blocks of the function; nil => external function.

	syntax ast.Node // AST node that represents the Function.

	// The following fields are set transiently during building,
	// then cleared.
	currentBlock *BasicBlock // Where to add instructions.
}

// Signature represents a function or method signature.
type Signature struct {
	Params  []*Parameter // Parameters from left to right; or nil
	Results []*Parameter // Return values from left to right; or nil
}

// Parameter represents an input parameter of a function or method.
type Parameter struct {
	node
	name  string // Name of parameter.
	Value Value  // Default value of parameter or nil.

	parent *Function // Function that the this parameter belongs.
}

// Const represents the value of a constant expression.
type Const struct {
	node
	Value string // Value of constant.
}

// Global is a named Value holding the address of a file-level
// variable.
type Global struct {
	node
	name  string   // Name of variable.
	Value ast.Expr // Value of global variable.
}

// Var represents a variable declaration.
type Var struct {
	node
	name  string // Name of variable.
	Value Value  // Value of variable
}

// Call instruction represents a function or method call.
type Call struct {
	node
	Parent   *Function // Function that Call is inside.
	Function *Function // The function that is being called.
	Args     []Value   // The call function parameters.
}

// BinOp instruction yields the result of binary operation Left Op Right.
type BinOp struct {
	node
	Op    string // Operator.
	Left  Value  // Left operand.
	Right Value  // Right operand.
}

// node is a mix-in embedded by all IR nodes to provide source code information.
//
// Since node is embedded by all IR nodes (Value's and Instruction's), these nodes
// implements the ast.Node interface which provide the original source code location.
type node struct {
	syntax ast.Node // Original code used to create the IR Value/Instruction.
}

// Pos implements ast.Node interface.
func (n node) Pos() ast.Position { return n.syntax.Pos() }

func (c *Const) value()         {}
func (c *Const) Name() string   { return fmt.Sprintf("%q", c.Value) }
func (c *Const) String() string { return c.Name() }

func (*Global) value()           {}
func (*Global) member()          {}
func (g *Global) Name() string   { return g.name }
func (g *Global) String() string { return g.Name() }

func (*Var) value()           {}
func (*Var) instr()           {}
func (v *Var) Name() string   { return v.name }
func (v *Var) String() string { return v.Value.String() }

func (*Parameter) value()           {}
func (p *Parameter) Name() string   { return p.name }
func (p *Parameter) String() string { return p.Name() }

func (*Call) instr()         {}
func (*Call) value()         {}
func (c *Call) Name() string { return "" }

func (*BinOp) instr()         {}
func (*BinOp) value()         {}
func (b *BinOp) Name() string { return b.String() }
func (b *BinOp) String() string {
	return fmt.Sprintf("%s %s %s", b.Left.Name(), b.Op, b.Right.Name())
}

func (*Function) member()        {}
func (m *Function) Name() string { return m.name }

func (*File) member()        {}
func (f *File) Name() string { return f.name }

func (*ExternalMember) member() {}
func (m *ExternalMember) Name() string {
	if m.Alias != "" {
		return m.Alias
	}

	return m.name
}

// Func returns the file-level function of the specified name or nil
// if not found.
//
// If the specified name is a imported function, Func will instantiate
// a new Function object using the path and name of imported function
// as a function name.
//
// nolint:funlen // There is no need to break this method.
func (f *File) Func(name string) *Function {
	fn, exists := f.Members[name]
	if !exists {
		return nil
	}

	switch fn := fn.(type) {
	case *Function:
		return fn
	case *ExternalMember:
		// TODO(matheus): Deal with empty ExternalMember.name fields.
		return &Function{
			name: fmt.Sprintf("%s.%s", fn.Path, fn.name),
			File: f,
		}
	default:
		panic(fmt.Sprintf("ir.File.Func: unexpected function member %T", fn))
	}
}

// ImportedPackage returns the importable package to a given name.
func (f *File) ImportedPackage(name string) *ExternalMember {
	return f.imported[name]
}

// lookup return the declared variable with the given name.
func (fn *Function) lookup(name string) *Var {
	return fn.Locals[name]
}
