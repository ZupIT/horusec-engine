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

	findings := Run(testProgram, rules).Value()

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

	findings := Run(testProgram, rules).Value()

	if len(findings) != 3000 {
		t.Fatal("Should find only 3000 finding")
	}

}
