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

package value

import (
	"strings"

	"github.com/ZupIT/horusec-engine/internal/ir"
	"github.com/ZupIT/horusec-engine/semantic/analysis"
)

// IsConst analyzer check if an ir.Value value is a constant value.
var IsConst analysis.AnalyzerValue = isConst{}

// Assert at compile time that Contains implements analysis.Analyzer interface.
var _ analysis.AnalyzerValue = &Contains{}

// Contains implements analysis.AnalyzerValue interface.
//
// Contains analyzer check if an ir.Value value contains a value
// informed on Contains.Values field.
type Contains struct {
	Values []string
}

// Run implements analysis.AnalyzerValue.Run.
//
// nolint: funlen,gocyclo // We need to do some type checking here.
func (a Contains) Run(v ir.Value) bool {
	switch value := v.(type) {
	case *ir.Const:
		constValue := strings.ToLower(value.Value)
		for _, v := range a.Values {
			if strings.EqualFold(constValue, strings.ToLower(v)) {
				return true
			}
		}
	case *ir.Var:
		return a.Run(value.Value)
	case *ir.Phi:
		return runPhiNode(value, a)
	case *ir.Object:
		return runValues(value.Values, a)
	}

	return false
}

// isConst implements analysis.AnalyzerValue interface.
//
// Read IsConst docs for more info.
type isConst struct{}

// Run check if a Value v is a constante value. If v is a variable or a phi
// value, Run will recursivily check the value of variable and the edges of
// phi node.
func (a isConst) Run(v ir.Value) bool {
	switch value := v.(type) {
	case *ir.Const:
		return true
	case *ir.Var:
		return a.Run(value.Value)
	case *ir.Phi:
		return runPhiNode(value, a)
	case *ir.Object:
		return runValues(value.Values, a)
	}

	return false
}

func runPhiNode(phi *ir.Phi, analyzer analysis.AnalyzerValue) bool {
	values := make([]ir.Value, 0, len(phi.Edges))
	for _, edge := range phi.Edges {
		values = append(values, edge)
	}
	return runValues(values, analyzer)
}

func runValues(values []ir.Value, analyzer analysis.AnalyzerValue) bool {
	for _, v := range values {
		if !analyzer.Run(v) {
			return false
		}
	}
	return true
}
