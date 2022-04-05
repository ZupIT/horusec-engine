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
// An Instruction that defines a value (e.g BinOp) also implements the Value interface,
// an Instruction that only has an effect (e.g Return) does not.
type Instruction interface {
	instr()

	// String returns the disassembled form of this value.
	//
	// Examples of Instructions that are Values:
	//       "x + y"     (BinOp)
	//       "len([])"   (Call)
	//
	// Examples of Instructions that are not Values:
	//       "return x"  (Return)
	String() string
}

// File represents a single file converted to IR representation.
type File struct {
	name    string            // Name of file.
	Members map[string]Member // All file members keyed by name.

	expressions []ast.Expr                 // Top level file expressions.
	imported    map[string]*ExternalMember // All importable packages, keyed by import path.
	syntax      ast.Node                   // AST representation of a file, contains all syntax nodes of the file
}

// ExternalMember represents a member that is declared outside the file that is being used.
//
// ExternalMember is created from imports of a single file.
//
// The ExternalMember implements Member interface.
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
	Preds   []*BasicBlock // Predecessors blocks.
	Succs   []*BasicBlock // Successors blocks.

	locals map[string]*Var // Local variables declared on this block.
}

// Function represents a function or method with the parameters and signature.
//
// The Function implements Member and Value interfaces.
type Function struct {
	name      string        // Function name.
	File      *File         // File that this function belongs.
	Signature *Signature    // Function signature.
	Blocks    []*BasicBlock // Basic blocks of the function; nil => external function.
	AnonFuncs []*Function   // Anonymous functions directly beneath this one.
	Locals    []*Var        // Local variables declared on this function.

	parent  *Function       // enclosing function if anonymous; nil if global.
	syntax  ast.Node        // AST node that represents the Function.
	nLocals int             // Number of local variables declared on this function.
	phis    map[string]*Phi // Phi values already computed to variable names.

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
//
// The Parameter implements Value interface.
type Parameter struct {
	node
	name     string // Name of parameter.
	Value    Value  // Default value of parameter or nil.
	Receiver bool   // True if parameter is a receiver of a struct method.

	parent *Function // Function that the this parameter belongs.
}

// Const represents the value of a constant expression.
//
// The Const implements Value interface.
type Const struct {
	node
	Value string // Value of constant.
}

// Global is a named Value holding the address of a file-level
// variable.
//
// Example printed form:
// var a
//
// The Global implements Value and Member interfaces.
type Global struct {
	node
	name  string   // Name of variable.
	Value ast.Expr // Value of global variable.
}

// Var represents a variable declaration.
//
// Example printed form:
// a = "hello"
// b = "20" // TODO(matheus): We should try to add a simple type checker to print the value correctly.
//
// The Var implements Value and Instruction interfaces.
type Var struct {
	node
	Label string // Name of variable on source code.
	Value Value  // Value of variable

	name  string      // Name of variable.
	block *BasicBlock // Basic block where this variable was defined.
}

// Template represents a template string.
//
// Example printed form:
// "template string: ${someVar}"
//
// The Template implements Value interface.
type Template struct {
	node
	Value string  // Template string.
	Subs  []Value // Substitution values.
}

// Selector value represents a parsed *ast.SelectorExpr which holds
// the field being acessed and the value used to access this field.
//
// Example printed form:
//	%t0 = a.b.c()
//	%t1 = %t0.d.e()
//
// The Selector implements Value interface.
type Selector struct {
	node
	Field *Var  // Field selector being accessed.
	Value Value // Value used to access Field.
}

// Phi instruction represents an SSA φ-node, which combines values
// that differ across incoming control-flow edges and yields a new
// value.  Within a block, all φ-nodes must appear before all non-φ
// nodes.
//
// Example printed form:
// 	t2 = phi [0: t0, 1: t1]
type Phi struct {
	Comment string // Optional label; no semantic significance.
	Edges   []*Var // Variables that differ across control-flow edges.
}

// Call instruction represents a function or method call.
//
// Function call arguments will never be another function call.
// If the Call argument is another function call, a temporary
// variable is created by associating the function call with its
// value, and the temporary variable created is used as an argument
// to Call.
//
// This code:
// 	foo(bar())
// Generate this IR:
// 	%t0 = bar()
//	foo(%t0)
//
// Example printed form:
// 	foo()
// 	bar.foo(10, 20)
//
// The Call implements Value and Instruction interfaces.
type Call struct {
	node
	Parent   *Function // Function that Call is inside.
	Function *Function // The function that is being called.
	Args     []Value   // The call function parameters.
}

// BinOp instruction yields the result of binary operation Left Op Right.
//
// Example printed form:
// a + b
//
// The BinOp implements Value and Instruction interfaces.
type BinOp struct {
	node
	Op    string // Operator.
	Left  Value  // Left operand.
	Right Value  // Right operand.
}

// Return instruction contains the return values of a function in some BasicBlock.
//
// In theory len(Results) is always be equal to the number of results in the
// function's signature, but since some languages don't declare the return values
// of a function, these values could be different.
//
// Return **must** be the last instruction of its containing BasicBlock.
//
// Example printed form:
// 	return
// 	return "x", 10
//
// The Return implements Instruction interface.
type Return struct {
	node
	Results []Value
}

// Struct is an IR instruction that represents a struct or a class with your
// fields and methods.
//
// Example printed form:
// type   ExampleClass
//   method(ExampleClass) exampleMethod(self)
//
// The Struct implements Value and Member interfaces.
type Struct struct {
	node
	name    string      // Struct name.
	Fields  []Value     // Struct attributes.
	Methods []*Function // Struct methods.
}

// Closure instruction yields a closure value whose code is Fn.
//
// Example printed form:
//  # Where sum is the closure function created inside a normal function.
// 	sum = make closure sum
//
//  # Where f1 is the function parent of closure and $1 is the number of closures inside parent.
// 	fs.readFile(path, f1$1)
//
// The Closure implements Instruction interface.
type Closure struct {
	Fn *Function // Closure function.
}

// If instruction transfers control to one of the two successors of
// its owning block, depending on the boolean Cond: the first if true,
// the second if false.
//
// An If instruction must be the last instruction of its containing
// BasicBlock.
//
// Example printed form (Consider 1 as if.them block and 3 as if.done block)
// 	if a > b goto 1 else 3
//
// The If implements Instruction interface.
type If struct {
	block *BasicBlock
	Cond  Value
}

// Jump instruction transfers control to the sole successor of its
// owning block.
//
// A Jump must be the last instruction of its containing BasicBlock.
//
// Example printed form (Consider 5 as a if.done block):
// 	jump 5
//
// The Jump implements Instruction interface.
type Jump struct {
	block *BasicBlock
}

// Object is an IR Value that represents arrays, constructors and hashmaps.
//
// Example printed form:
// %t0 = array["a","b","c"]
// %t1 = constructor(ExClass)
//
// The Object implements the Value interface.
// TODO: in the future it will be necessary to improve the hashmap print
type Object struct {
	node
	Type    Value
	Comment string
	Values  []Value
}

// HashMap is an IR Value that represents a hashmap.
//
// Example printed form:
// %t0 = hashmap[{"k": "v"},{"k2": "v2"}]
//
// The HashMap implements the Value interface.
type HashMap struct {
	node
	Key   Value
	Value Value
}

// Lookup instruction yields element from Index of Object.
//
// Example printed form:
// %t0 = array["a","b","c"]
// %t1 = "0"
// %t2 = %t0[%t1]
//
// The Lookup implements the Value and Instruction interfaces.
type Lookup struct {
	node
	Object Value // object being subscript by the index
	Index  Value // value used to subscript the object
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

// ------------------------------------------------------------------------
// Implementations of Member, Value and Instruction interfaces.
// ------------------------------------------------------------------------

func (c *Const) value()         {}
func (c *Const) Name() string   { return fmt.Sprintf("%q", c.Value) }
func (c *Const) String() string { return c.Name() }

func (*Global) value()           {}
func (*Global) member()          {}
func (g *Global) Name() string   { return g.name }
func (g *Global) String() string { return g.Name() }

func (*Var) value()         {}
func (*Var) instr()         {}
func (v *Var) Name() string { return v.name }
func (v *Var) String() string {
	// since we can return a nil for unsupported nodes this is necessary.
	if v.Value == nil {
		return "nil"
	}

	return v.Value.String()
}

func (*Parameter) value()           {}
func (p *Parameter) Name() string   { return p.name }
func (p *Parameter) String() string { return p.Name() }

func (*Template) value()           {}
func (t *Template) Name() string   { return t.String() }
func (t *Template) String() string { return fmt.Sprintf("%q", t.Value) }

func (*Call) instr()         {}
func (*Call) value()         {}
func (c *Call) Name() string { return "" }

func (*BinOp) instr()         {}
func (*BinOp) value()         {}
func (b *BinOp) Name() string { return b.String() }
func (b *BinOp) String() string {
	return fmt.Sprintf("%s %s %s", b.Left.Name(), b.Op, b.Right.Name())
}

func (*Return) instr() {}

func (*Closure) instr()           {}
func (*Closure) value()           {}
func (c *Closure) Name() string   { return c.Fn.Name() }
func (c *Closure) String() string { return fmt.Sprintf("make closure %s ", c.Name()) }

func (*If) instr() {}

func (*Jump) instr() {}

func (*Phi) instr()       {}
func (*Phi) value()       {}
func (*Phi) Name() string { return "" }

func (*Function) value()            {}
func (*Function) member()           {}
func (fn *Function) Name() string   { return fn.name }
func (fn *Function) String() string { return "" }

func (*File) member()        {}
func (f *File) Name() string { return f.name }

func (*ExternalMember) value()           {}
func (*ExternalMember) member()          {}
func (m *ExternalMember) Name() string   { return m.Path }
func (m *ExternalMember) String() string { return m.Name() }

func (*Struct) value()           {}
func (*Struct) member()          {}
func (s *Struct) Name() string   { return s.name }
func (s *Struct) String() string { return fmt.Sprintf("type %s struct", s.Name()) }

func (*Object) value() {}
func (o *Object) Name() string {
	if o.Type != nil {
		return o.Type.Name()
	}
	return ""
}

func (*HashMap) value()         {}
func (h *HashMap) Name() string { return h.String() }
func (h *HashMap) String() string {
	return fmt.Sprintf("{%s: %s}", h.Key.String(), h.Value.String())
}

func (*Lookup) value()           {}
func (*Lookup) instr()           {}
func (l *Lookup) Name() string   { return l.Object.Name() }
func (l *Lookup) String() string { return fmt.Sprintf("%s[%s]", l.Name(), l.Index.Name()) }

func (*Selector) value() {}

// nolint: funlen // This linter considers lines commented out which makes no sense.
func (s *Selector) Name() string {
	// If the value of Selector is a variable we need to check if
	// the value of this variable is a object declaration, if it is
	// we use the object name (if its not empty) instead the variable
	// name, since the field being acessed is a value from a object.
	//
	// TODO(matheus): We should find a better approach to deal with this.

	var name string

	if v, ok := s.Value.(*Var); ok {
		if obj, ok := v.Value.(*Object); ok {
			name = obj.Name()
		}
	}

	if name == "" {
		name = s.Value.Name()
	}

	return fmt.Sprintf("%s.%s", name, s.Field.Name())
}
func (s *Selector) String() string { return s.Name() }

// ------------------------------------------------------------------------

// ImportName return the name used on source code from a imported package.
// Could be an  alias or the real name.
func (m *ExternalMember) ImportName() string {
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
		panic(fmt.Sprintf("ir.File.Func: unexpected function member %s: %T", name, fn))
	}
}

// ImportedPackage returns the importable package to a given name.
func (f *File) ImportedPackage(name string) *ExternalMember {
	return f.imported[name]
}

// lookup return the declared variable in function with the given name.
//
// If variable is not declared on the current block of fn, lookup will
// recursively search on predecessors blocks of fn.currentBlock and will
// return a φ(phi)-node with the possible values to the given variable name.
//
// If fn.currentBlock has a single predecessor we just search on this block
// and return the variable if exists.
//
// nolint:funlen // The function is simple enought to not split
func (fn *Function) lookup(name string) Value {
	// Check if variable exists in the current basic block,
	// if exists we return a pointer to this variable.
	if v, exists := fn.currentBlock.locals[name]; exists {
		return v
	}

	// Check if the phi value was already computed, if yes, just returned it.
	if phi, exists := fn.phis[name]; exists {
		return phi
	}

	phi := &Phi{
		Comment: name,
		Edges:   make([]*Var, 0),
	}

	// Try to compute the phi values recursively in all basic block predecessors and cache it.
	fn.recursivelyLoopkup(name, fn.currentBlock, phi, make(map[int]bool, len(fn.currentBlock.Preds)))

	if len(phi.Edges) == 0 {
		// We don't find any variable at any block with the given name, so return nil.
		return nil
	} else if len(phi.Edges) == 1 {
		// This means that the the code has just one variable declaration with the given
		// name, so we don't need create a phi value for this case, just return the declared
		// variable.
		return phi.Edges[0]
	}

	fn.phis[name] = phi

	return fn.addLocal(phi, nil)
}

// recursivelyLoopkup recursively search for a variable with the given name on predecessors
// blocks of the given basic block and return all founded variables.
func (fn *Function) recursivelyLoopkup(name string, block *BasicBlock, phi *Phi, visitedBlocks map[int]bool) {
	for _, block := range block.Preds {
		// Store already visited blocks to avoid endless recursion.
		if _, visited := visitedBlocks[block.Index]; visited {
			continue
		}
		visitedBlocks[block.Index] = true

		if v, exists := block.locals[name]; exists {
			phi.Edges = append(phi.Edges, v)
			continue
		}
		fn.recursivelyLoopkup(name, block, phi, visitedBlocks)
	}
}
