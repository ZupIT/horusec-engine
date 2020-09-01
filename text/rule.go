package text

import (
	"regexp"

	"github.com/ZupIT/horus-engine"
)

type MatchType int

const (
	Regular MatchType = iota
	NotMatch
	OrMatch
	AndMatch
)

type TextRule struct {
	engine.Metadata
	Type        MatchType
	Expressions []*regexp.Regexp
}

func (rule TextRule) IsFor(unitType engine.UnitType) bool {
	return engine.ProgramTextUnit == unitType
}
