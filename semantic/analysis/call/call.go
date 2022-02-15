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
	"github.com/ZupIT/horusec-engine/internal/ir"
	"github.com/ZupIT/horusec-engine/semantic/analysis"
)

// Assert at compile time that Analyzer implements analysis.Analyzer interface.
var _ analysis.Analyzer = &Analyzer{}

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
			if call, ok := instr.(*ir.Call); ok && a.isVulnerableCall(call) {
				pass.Report(analysis.NewIssue(pass.File.Name(), call))
			}
		}
	}
}

func (a *Analyzer) isVulnerableCall(call *ir.Call) bool {
	if call.Function.Name() != a.Name || !a.hasRequiredArgs(call) {
		return false
	}

	return !a.ArgValue.Run(call.Args[a.ArgsIndex-1])
}

func (a *Analyzer) hasRequiredArgs(call *ir.Call) bool {
	return len(call.Args) >= a.ArgsIndex
}