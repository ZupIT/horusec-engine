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
	"github.com/ZupIT/horusec-engine/internal/ast"
)

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
}

// Instruction is an IR instruction that computes a new Value or has some effect.
type Instruction interface {
	instr()
}

// BasicBlock represents an IR basic block.
type BasicBlock struct {
	Comment string        // Optional label; no semantic significance
	Instrs  []Instruction // Instructions in order.
}

// Function represents a function or method with the parameters and signature.
type Function struct {
	Name      string          // Function name.
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
	Name  string // Name of parameter.
	Value Value  // Default value of parameter or nil.

	parent *Function // Function that the this parameter belongs.
}

// Const represents the value of a constant expression.
type Const struct {
	Value string // Value of constant.
}

// Var represents a variable
type Var struct {
	name  string // Name of variable.
	Value Value  // Value of variable
}

// Call instruction represents a function or method call.
type Call struct {
	Parent   *Function // Function that Call is inside.
	Function *Function // The function that is being called.
	Args     []Value   // The call function parameters.
}

func (*Const) value() {}
func (*Var) value()   {}

func (*Call) instr() {}
