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

package call

import (
	"github.com/ZupIT/horusec-engine/internal/ast"
	"github.com/ZupIT/horusec-engine/internal/ir"
	"github.com/ZupIT/horusec-engine/semantic/analysis"
)

// Assert at compile time that Analyzer implements analysis.Analyzer interface.
var _ analysis.Analyzer = &Analyzer{}

const (
	// AllArguments represents that all call arguments would be analyzed using
	// the Analyzer.ArgValue analyzer
	AllArguments = -1

	// NoArguments represents that the function arguments will not be analyzed.
	NoArguments = -2
)

// Analyzer implements analysis.Analyzer interface.
//
// Call analyzer check if a function call match the configured
// options on fields.
type Analyzer struct {
	// Name is the name of function that should be analyzed.
	// If the function belongs a package, the name should contains
	// the all package name separed by dot e.g: `fs.readFile`
	Name string

	// ArgsIndex is the index of argument that should be analyzed.
	ArgsIndex int

	// ArgValue is the analyzer that will be used to validate the
	// function argument in the index informed in ArgsIndex.
	ArgValue analysis.AnalyzerValue
}

// Run implements analysis.Analyzer.Run.
func (a *Analyzer) Run(pass *analysis.Pass) {
	for _, block := range pass.Function.Blocks {
		for _, instr := range block.Instrs {
			if node, isVulnerable := a.checkInstruction(instr); isVulnerable {
				pass.Report(analysis.NewIssue(pass.File.Name(), node))
			}
		}
	}
}

// checkInstruction check if an instruction instr has a vulnerable call.
//
// If instruction is vulnerable, the return value will the respective ast.Node
// and a true value informing that the instruction is vulnerable, otherwise
// a nil ast.Node and false values are returned.
//
// If instruction value is a variable, checkInstruction will analyze the variable
// value to check if is vulnerable function call.
func (a *Analyzer) checkInstruction(instr ir.Instruction) (ast.Node, bool) {
	switch instr := instr.(type) {
	case *ir.Call:
		return instr, a.isVulnerableCall(instr)
	case *ir.Var:
		if value, ok := instr.Value.(ir.Instruction); ok {
			return a.checkInstruction(value)
		}
	}
	return nil, false
}

// nolint: funlen,gocyclo // There is no needed to break this function.
func (a *Analyzer) isVulnerableCall(call *ir.Call) bool {
	// Fast path, if function name don't match return false directly.
	if call.Function.Name() != a.Name {
		return false
	}

	// If the function arguments should not be checked, return true directly
	// since the function name is already matched.
	if a.ArgsIndex == NoArguments {
		return true
	}

	if a.ArgsIndex == AllArguments {
		for _, arg := range call.Args {
			if !a.ArgValue.Run(arg) {
				return true
			}
		}
		return false
	}

	return a.hasRequiredArgs(call) && !a.ArgValue.Run(call.Args[a.ArgsIndex-1])
}

func (a *Analyzer) hasRequiredArgs(call *ir.Call) bool {
	return len(call.Args) >= a.ArgsIndex
}
