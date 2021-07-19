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

package engine

import (
	"testing"
)

var TestUnitType UnitType = 999

type TestRule struct{}

func (rule TestRule) IsFor(unitType UnitType) bool {
	return TestUnitType == unitType
}

type TestUnit struct{}

func (unit TestUnit) Type() UnitType {
	return TestUnitType
}

func (unit TestUnit) Eval(rule Rule) []Finding {
	return []Finding{
		Finding{
			ID: "1",
		},
	}
}

func TestRunWithTextUnits(t *testing.T) {
	testProgram := []Unit{TestUnit{}}
	rules := []Rule{TestRule{}}

	findings := Run(testProgram, rules)

	if len(findings) < 1 || len(findings) > 1 {
		t.Fatal("Should find only 1 finding")
	}
}

func TestRunWith1000Units(t *testing.T) {
	rules := []Rule{TestRule{}, TestRule{}, TestRule{}}
	testProgram := []Unit{}

	for i := 0; i < 1000; i++ {
		testProgram = append(testProgram, TestUnit{})
	}

	findings := Run(testProgram, rules)

	if len(findings) != 3000 {
		t.Fatal("Should find only 3000 finding")
	}
}
